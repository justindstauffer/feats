package com.jstauff.feats.android.core.network

import com.jstauff.feats.android.BuildConfig
import com.jstauff.feats.android.core.storage.SecureTokenStore
import kotlinx.serialization.json.Json
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.create
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import retrofit2.http.Multipart
import retrofit2.http.Part
import retrofit2.http.Path
import retrofit2.http.POST

object ApiClient {
    @Volatile
    private var initialized = false

    private lateinit var tokenStore: SecureTokenStore
    private var accessToken: String? = null

    private val json = Json {
        ignoreUnknownKeys = true
        explicitNulls = false
    }

    private val authInterceptor = Interceptor { chain ->
        val original = chain.request()
        // Do NOT force Content-Type here. Retrofit already sets it correctly per
        // request: application/json for JSON bodies (via the serialization
        // converter) and multipart/form-data; boundary=... for image uploads.
        // Forcing application/json overwrote the multipart boundary, so the
        // backend's FormFile("image") could not parse the body and every image
        // upload failed with 400 "No image file provided".
        val builder = original.newBuilder()
            .header("Cache-Control", "no-cache")
            .header("Pragma", "no-cache")

        accessToken?.let { token ->
            builder.header("Authorization", "Bearer $token")
        }

        val request: Request = builder.build()
        chain.proceed(request)
    }

    private val httpClient: OkHttpClient by lazy {
        OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .addInterceptor(HttpLoggingInterceptor().apply {
                level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BODY else HttpLoggingInterceptor.Level.NONE
            })
            .build()
    }

    private val retrofit: Retrofit by lazy {
        Retrofit.Builder()
            .baseUrl(BuildConfig.API_BASE_URL)
            .client(httpClient)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
    }

    val api: FeatsApi by lazy { retrofit.create() }
    private val uploadApi: UploadApi by lazy { retrofit.create() }

    fun initialize(secureTokenStore: SecureTokenStore) {
        if (initialized) return
        tokenStore = secureTokenStore
        initialized = true
    }

    fun setAccessToken(token: String?) {
        accessToken = token
    }

    fun getAccessToken(): String? = accessToken

    fun saveRefreshToken(token: String) {
        tokenStore.saveRefreshToken(token)
    }

    fun getRefreshToken(): String? = tokenStore.getRefreshToken()

    fun clearSession() {
        accessToken = null
        tokenStore.clear()
    }

    fun webSocketUrl(): String? {
        val token = accessToken ?: return null
        val base = BuildConfig.WS_BASE_URL.toHttpUrlOrNull() ?: return null
        return base.newBuilder()
            .addQueryParameter("token", token)
            .build()
            .toString()
    }

    fun imageUrl(imageId: String): String? {
        val apiBase = BuildConfig.API_BASE_URL.toHttpUrlOrNull() ?: return null
        val root = apiBase.newBuilder()
            .encodedPath("/")
            .build()
        return root.newBuilder()
            .addPathSegment("images")
            .addPathSegment(imageId)
            .build()
            .toString()
    }

    suspend fun uploadPostImage(groupId: String, postId: String, imageBytes: ByteArray, filename: String): Boolean {
        val part = MultipartBody.Part.createFormData(
            name = "image",
            filename = filename,
            body = imageBytes.toRequestBody("image/jpeg".toMediaTypeOrNull())
        )
        val response = uploadApi.uploadPostImage(groupId = groupId, postId = postId, image = part)
        return response.data != null && response.error == null
    }

    private interface UploadApi {
        @Multipart
        @POST("groups/{groupId}/posts/{postId}/images")
        suspend fun uploadPostImage(
            @Path("groupId") groupId: String,
            @Path("postId") postId: String,
            @Part image: MultipartBody.Part
        ): com.jstauff.feats.android.core.network.dto.ApiResponse<Map<String, String>>
    }
}
