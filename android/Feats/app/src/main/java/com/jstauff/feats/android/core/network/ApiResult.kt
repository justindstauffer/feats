package com.jstauff.feats.android.core.network

import com.jstauff.feats.android.core.network.dto.ApiError
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import retrofit2.HttpException
import java.io.IOException

/**
 * Minimal view of the backend envelope used only to recover [ApiError] from a
 * non-2xx body. Retrofit throws [HttpException] before the typed `ApiResponse<T>`
 * converter runs, so the server's message is otherwise unreachable.
 */
@Serializable
private data class ErrorEnvelope(val error: ApiError? = null)

sealed interface ApiResult<out T> {
    data class Success<T>(val value: T) : ApiResult<T>

    /**
     * @param code backend error code (see SPECIFICATION.md), or null for transport failures.
     * @param httpStatus null when the request never reached the server.
     */
    data class Failure(
        val message: String,
        val code: String? = null,
        val httpStatus: Int? = null
    ) : ApiResult<Nothing>

    val isUnauthorized: Boolean
        get() = this is Failure && httpStatus == 401
}

private val errorJson = Json { ignoreUnknownKeys = true }

/**
 * Runs [block] and normalises the three failure shapes — server error envelope,
 * transport error, and unexpected exception — into [ApiResult.Failure].
 *
 * Retrofit's suspend functions already dispatch on OkHttp's dispatcher, so no
 * `withContext(Dispatchers.IO)` wrapper is needed here.
 */
suspend fun <T> apiCall(block: suspend () -> T): ApiResult<T> = try {
    ApiResult.Success(block())
} catch (e: HttpException) {
    // errorBody().string() consumes the stream — parse exactly once.
    val serverError = e.errorEnvelope()?.error
    ApiResult.Failure(
        message = serverError?.message?.takeIf { it.isNotBlank() } ?: defaultMessageFor(e.code()),
        code = serverError?.code,
        httpStatus = e.code()
    )
} catch (e: IOException) {
    ApiResult.Failure("Can't reach Feats. Check your connection and try again.")
} catch (e: Exception) {
    ApiResult.Failure(e.message ?: "Something went wrong.")
}

private fun HttpException.errorEnvelope(): ErrorEnvelope? = try {
    response()?.errorBody()?.string()
        ?.takeIf { it.isNotBlank() }
        ?.let { errorJson.decodeFromString<ErrorEnvelope>(it) }
} catch (e: Exception) {
    null
}

private fun defaultMessageFor(status: Int): String = when (status) {
    401 -> "Your session expired. Please sign in again."
    403 -> "You don't have access to this."
    404 -> "That's no longer available."
    429 -> "Too many requests. Give it a moment and try again."
    in 500..599 -> "Feats is having trouble right now. Try again shortly."
    else -> "Something went wrong."
}
