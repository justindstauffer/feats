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
import com.jstauff.feats.android.core.network.dto.CreateGoalRequest
import com.jstauff.feats.android.core.network.dto.CreateGroupRequest
import com.jstauff.feats.android.core.network.dto.CreatePostRequest
import com.jstauff.feats.android.core.network.dto.UpdateGoalRequest
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.network.dto.CreateGroupInviteRequest
import com.jstauff.feats.android.core.network.dto.GroupDto
import com.jstauff.feats.android.core.network.dto.GroupInviteDto
import com.jstauff.feats.android.core.network.dto.GroupMemberDto
import com.jstauff.feats.android.core.network.dto.LoginPayload
import com.jstauff.feats.android.core.network.dto.UpdateGroupRequest
import com.jstauff.feats.android.core.network.dto.UpdateMemberRequest
import com.jstauff.feats.android.core.network.dto.LoginRequest
import com.jstauff.feats.android.core.network.dto.PaginatedResponse
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.RegisterDeviceRequest
import com.jstauff.feats.android.core.network.dto.ReactionDto
import com.jstauff.feats.android.core.network.dto.ReactionsPayloadDto
import com.jstauff.feats.android.core.network.dto.RedeemInviteRequest
import com.jstauff.feats.android.core.network.dto.RefreshRequest
import com.jstauff.feats.android.core.network.dto.RegisterRequest
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.network.dto.TokenPair
import com.jstauff.feats.android.core.network.dto.UpdateCommentRequest
import com.jstauff.feats.android.core.network.dto.UpdatePostRequest
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

    @POST("auth/register")
    suspend fun register(@Body request: RegisterRequest): ApiResponse<LoginPayload>

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

    @POST("groups/{groupId}/leave")
    suspend fun leaveGroup(@Path("groupId") groupId: String): ApiResponse<Map<String, String>>

    @GET("groups/{groupId}/invites")
    suspend fun groupInvites(@Path("groupId") groupId: String): ApiResponse<List<GroupInviteDto>>

    @POST("groups/{groupId}/invites")
    suspend fun createGroupInvite(
        @Path("groupId") groupId: String,
        @Body request: CreateGroupInviteRequest
    ): ApiResponse<GroupInviteDto>

    @DELETE("groups/{groupId}/invites/{inviteId}")
    suspend fun revokeGroupInvite(
        @Path("groupId") groupId: String,
        @Path("inviteId") inviteId: String
    ): ApiResponse<Map<String, String>>

    @GET("groups/{groupId}/members")
    suspend fun groupMembers(@Path("groupId") groupId: String): ApiResponse<List<GroupMemberDto>>

    @PUT("groups/{groupId}")
    suspend fun updateGroup(
        @Path("groupId") groupId: String,
        @Body request: UpdateGroupRequest
    ): ApiResponse<GroupDto>

    @DELETE("groups/{groupId}")
    suspend fun deleteGroup(@Path("groupId") groupId: String): ApiResponse<Map<String, String>>

    @PUT("groups/{groupId}/members/{userId}")
    suspend fun updateMember(
        @Path("groupId") groupId: String,
        @Path("userId") userId: String,
        @Body request: UpdateMemberRequest
    ): ApiResponse<Map<String, String>>

    @DELETE("groups/{groupId}/members/{userId}")
    suspend fun removeMember(
        @Path("groupId") groupId: String,
        @Path("userId") userId: String
    ): ApiResponse<Map<String, String>>

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

    @POST("groups/{groupId}/goals")
    suspend fun createGoal(
        @Path("groupId") groupId: String,
        @Body request: CreateGoalRequest
    ): ApiResponse<GoalDto>

    @PUT("groups/{groupId}/goals/{goalId}")
    suspend fun updateGoal(
        @Path("groupId") groupId: String,
        @Path("goalId") goalId: String,
        @Body request: UpdateGoalRequest
    ): ApiResponse<GoalDto>

    @DELETE("groups/{groupId}/goals/{goalId}")
    suspend fun deleteGoal(
        @Path("groupId") groupId: String,
        @Path("goalId") goalId: String
    ): ApiResponse<Map<String, String>>

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

    @PUT("groups/{groupId}/posts/{postId}")
    suspend fun updatePost(
        @Path("groupId") groupId: String,
        @Path("postId") postId: String,
        @Body request: UpdatePostRequest
    ): ApiResponse<PostDto>

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

    @PUT("groups/{groupId}/comments/{commentId}")
    suspend fun updateComment(
        @Path("groupId") groupId: String,
        @Path("commentId") commentId: String,
        @Body request: UpdateCommentRequest
    ): ApiResponse<CommentDto>

    @DELETE("groups/{groupId}/comments/{commentId}")
    suspend fun deleteComment(
        @Path("groupId") groupId: String,
        @Path("commentId") commentId: String
    ): ApiResponse<Map<String, String>>
}
