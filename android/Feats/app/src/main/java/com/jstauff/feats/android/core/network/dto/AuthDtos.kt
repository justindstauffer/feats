package com.jstauff.feats.android.core.network.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class ApiError(
    val code: String,
    val message: String
)

@Serializable
data class ApiResponse<T>(
    val data: T? = null,
    val error: ApiError? = null
)

@Serializable
data class LoginRequest(
    val email: String,
    val password: String
)

@Serializable
data class RegisterRequest(
    val email: String,
    val password: String,
    val name: String,
    @SerialName("invite_code") val inviteCode: String
)

@Serializable
data class RefreshRequest(
    @SerialName("refresh_token") val refreshToken: String
)

@Serializable
data class TokenPair(
    @SerialName("access_token") val accessToken: String,
    @SerialName("refresh_token") val refreshToken: String,
    @SerialName("expires_at") val expiresAt: String
)

@Serializable
data class LoginPayload(
    val tokens: TokenPair,
    val user: UserDto? = null
)

@Serializable
data class UserDto(
    val id: String,
    val email: String,
    val name: String,
    @SerialName("profile_picture") val profilePicture: String? = null,
    val bio: String? = null,
    val role: String? = null
)

@Serializable
data class RegisterDeviceRequest(
    val token: String,
    val platform: String
)

@Serializable
data class UpdateUserRequest(
    val name: String? = null,
    val bio: String? = null,
    @SerialName("profile_picture") val profilePicture: String? = null
)

@Serializable
data class ChangePasswordRequest(
    @SerialName("current_password") val currentPassword: String,
    @SerialName("new_password") val newPassword: String
)
