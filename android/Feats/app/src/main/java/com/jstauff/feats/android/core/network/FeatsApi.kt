package com.jstauff.feats.android.core.network

import com.jstauff.feats.android.core.network.dto.ApiResponse
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.AddReactionRequest
import com.jstauff.feats.android.core.network.dto.BetaInviteDto
import com.jstauff.feats.android.core.network.dto.ChallengeDto
import com.jstauff.feats.android.core.network.dto.ChangePasswordRequest
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.CreateChallengeRequest
import com.jstauff.feats.android.core.network.dto.CreateBetaInviteRequest
import com.jstauff.feats.android.core.network.dto.CreateCommentRequest
import com.jstauff.feats.android.core.network.dto.CreateGroupRequest
import com.jstauff.feats.android.core.network.dto.CreatePostRequest
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.network.dto.GroupDto
import com.jstauff.feats.android.core.network.dto.LoginPayload
import com.jstauff.feats.android.core.network.dto.LoginRequest
import com.jstauff.feats.android.core.network.dto.PaginatedResponse
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.RegisterDeviceRequest
import com.jstauff.feats.android.core.network.dto.ReactionDto
import com.jstauff.feats.android.core.network.dto.ReactionsPayloadDto
import com.jstauff.feats.android.core.network.dto.RedeemInviteRequest
import com.jstauff.feats.android.core.network.dto.RefreshRequest
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.network.dto.TokenPair
import com.jstauff.feats.android.core.network.dto.UpdateUserRequest
import com.jstauff.feats.android.core.network.dto.UserDto
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.HTTP
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.PUT
import retrofit2.http.Query

interface FeatsApi {
    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): ApiResponse<LoginPayload>

    @POST("auth/refresh")
    suspend fun refresh(@Body request: RefreshRequest): ApiResponse<TokenPair>

    @POST("devices")
    suspend fun registerDevice(@Body request: RegisterDeviceRequest): ApiResponse<Map<String, String>>

    @HTTP(method = "DELETE", path = "devices", hasBody = true)
    suspend fun unregisterDevice(@Body request: RegisterDeviceRequest): ApiResponse<Map<String, String>>

    @GET("users/me")
    suspend fun me(): ApiResponse<UserDto>

    @PUT("users/me")
    suspend fun updateMe(@Body request: UpdateUserRequest): ApiResponse<UserDto>

    @POST("auth/password/change")
    suspend fun changePassword(@Body request: ChangePasswordRequest): ApiResponse<Map<String, String>>

    @GET("groups")
    suspend fun groups(): ApiResponse<List<GroupDto>>

    @POST("groups")
    suspend fun createGroup(@Body request: CreateGroupRequest): ApiResponse<GroupDto>

    @POST("invites/redeem")
    suspend fun redeemInvite(@Body request: RedeemInviteRequest): ApiResponse<GroupDto>

    @GET("groups/{groupId}/activities")
    suspend fun activities(
        @Path("groupId") groupId: String
    ): ApiResponse<List<ActivityTypeDto>>

    @GET("groups/{groupId}/posts")
    suspend fun groupPosts(
        @Path("groupId") groupId: String,
        @Query("page") page: Int,
        @Query("per_page") perPage: Int
    ): PaginatedResponse<List<PostDto>>

    @POST("groups/{groupId}/posts")
    suspend fun createPost(
        @Path("groupId") groupId: String,
        @Body request: CreatePostRequest
    ): ApiResponse<PostDto>

    @GET("groups/{groupId}/streaks/leaderboard")
    suspend fun leaderboard(
        @Path("groupId") groupId: String
    ): ApiResponse<List<StreakDto>>

    @GET("groups/{groupId}/users/{userId}/streak")
    suspend fun userStreak(
        @Path("groupId") groupId: String,
        @Path("userId") userId: String
    ): ApiResponse<StreakDto>

    @GET("groups/{groupId}/users/{userId}/goals")
    suspend fun userGoals(
        @Path("groupId") groupId: String,
        @Path("userId") userId: String
    ): ApiResponse<List<GoalDto>>

    @GET("groups/{groupId}/challenges")
    suspend fun challenges(
        @Path("groupId") groupId: String,
        @Query("include_expired") includeExpired: Boolean = true
    ): ApiResponse<List<ChallengeDto>>

    @POST("groups/{groupId}/challenges")
    suspend fun createChallenge(
        @Path("groupId") groupId: String,
        @Body request: CreateChallengeRequest
    ): ApiResponse<ChallengeDto>

    @POST("groups/{groupId}/challenges/{challengeId}/join")
    suspend fun joinChallenge(
        @Path("groupId") groupId: String,
        @Path("challengeId") challengeId: String
    ): ApiResponse<Map<String, String>>

    @DELETE("groups/{groupId}/challenges/{challengeId}/leave")
    suspend fun leaveChallenge(
        @Path("groupId") groupId: String,
        @Path("challengeId") challengeId: String
    ): ApiResponse<Map<String, String>>

    @GET("admin/beta-invites")
    suspend fun listBetaInvites(): ApiResponse<List<BetaInviteDto>>

    @POST("admin/beta-invites")
    suspend fun createBetaInvite(@Body request: CreateBetaInviteRequest): ApiResponse<BetaInviteDto>

    @DELETE("admin/beta-invites/{inviteId}")
    suspend fun deleteBetaInvite(@Path("inviteId") inviteId: String): ApiResponse<Map<String, String>>

    @GET("groups/{groupId}/posts/{postId}")
    suspend fun groupPostById(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String
    ): ApiResponse<PostDto>

    @DELETE("groups/{groupId}/posts/{postId}")
    suspend fun deletePost(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String
    ): ApiResponse<Map<String, String>>

    @GET("groups/{groupId}/posts/{postId}/reactions")
    suspend fun reactions(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String
    ): ApiResponse<ReactionsPayloadDto>

    @POST("groups/{groupId}/posts/{postId}/reactions")
    suspend fun addReaction(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String,
        @Body request: AddReactionRequest
    ): ApiResponse<ReactionDto>

    @DELETE("groups/{groupId}/posts/{postId}/reactions")
    suspend fun removeReaction(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String
    ): ApiResponse<Map<String, String>>

    @GET("groups/{groupId}/posts/{postId}/comments")
    suspend fun comments(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String
    ): ApiResponse<List<CommentDto>>

    @POST("groups/{groupId}/posts/{postId}/comments")
    suspend fun createComment(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String,
        @Body request: CreateCommentRequest
    ): ApiResponse<CommentDto>
}
