package com.jstauff.feats.android.core.network.dto

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class PaginatedResponse<T>(
    val data: T? = null,
    val error: ApiError? = null,
    val pagination: Pagination? = null
)

@Serializable
data class Pagination(
    val page: Int,
    @SerialName("per_page") val perPage: Int,
    val total: Int,
    @SerialName("total_pages") val totalPages: Int
)

@Serializable
data class GroupDto(
    val id: String,
    val name: String,
    val description: String? = null,
    @SerialName("created_by") val createdBy: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String
)

@Serializable
data class ActivityTypeDto(
    val id: String,
    val name: String,
    val icon: String? = null,
    @SerialName("is_system") val isSystem: Boolean = true,
    @SerialName("created_at") val createdAt: String? = null
)

@Serializable
data class ReactionDto(
    val id: String,
    @SerialName("user_id") val userId: String,
    @SerialName("post_id") val postId: String,
    @SerialName("reaction_type") val reactionType: Int,
    @SerialName("created_at") val createdAt: String
)

@Serializable
data class PostImageDto(
    val id: String,
    @SerialName("post_id") val postId: String,
    @SerialName("display_order") val displayOrder: Int,
    @SerialName("created_at") val createdAt: String
)

@Serializable
data class PostDto(
    val id: String,
    @SerialName("user_id") val userId: String,
    @SerialName("activity_type_id") val activityTypeId: String,
    val description: String? = null,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String,
    val user: UserDto? = null,
    @SerialName("activity_type") val activityType: ActivityTypeDto? = null,
    val images: List<PostImageDto>? = null,
    val reactions: List<ReactionDto>? = null,
    @SerialName("comment_count") val commentCount: Int? = null
)

@Serializable
data class ReactionSummaryDto(
    val type: Int,
    val emoji: String,
    val count: Int
)

@Serializable
data class ReactionsPayloadDto(
    val summary: List<ReactionSummaryDto>? = null,
    val reactions: List<ReactionDto>? = null
)

@Serializable
data class AddReactionRequest(
    @SerialName("reaction_type") val reactionType: Int
)

@Serializable
data class CommentDto(
    val id: String,
    @SerialName("post_id") val postId: String,
    @SerialName("user_id") val userId: String,
    @SerialName("parent_id") val parentId: String? = null,
    val content: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String,
    val user: UserDto? = null,
    val replies: List<CommentDto>? = null
)

@Serializable
data class CreateCommentRequest(
    val content: String,
    @SerialName("parent_id") val parentId: String? = null
)

@Serializable
data class CreateGroupRequest(
    val name: String,
    val description: String? = null
)

@Serializable
data class RedeemInviteRequest(
    val code: String
)

@Serializable
data class GroupInviteDto(
    val id: String,
    @SerialName("group_id") val groupId: String,
    val code: String,
    @SerialName("created_by") val createdBy: String,
    @SerialName("expires_at") val expiresAt: String,
    @SerialName("max_uses") val maxUses: Int,
    @SerialName("use_count") val useCount: Int,
    @SerialName("created_at") val createdAt: String? = null,
    val creator: UserDto? = null
)

@Serializable
data class CreateGroupInviteRequest(
    @SerialName("max_uses") val maxUses: Int = 1,
    @SerialName("expires_in") val expiresIn: Int = 168
)

@Serializable
data class CreatePostRequest(
    @SerialName("activity_type_id") val activityTypeId: String,
    val description: String? = null
)

@Serializable
data class UpdatePostRequest(
    val description: String? = null
)

@Serializable
data class UpdateCommentRequest(
    val content: String
)

@Serializable
data class ChallengeParticipantDto(
    val id: String,
    @SerialName("challenge_id") val challengeId: String,
    @SerialName("user_id") val userId: String,
    val progress: Int,
    @SerialName("completed_at") val completedAt: String? = null,
    @SerialName("joined_at") val joinedAt: String,
    val user: UserDto? = null
)

@Serializable
data class ChallengeDto(
    val id: String,
    @SerialName("created_by") val createdBy: String,
    val title: String,
    val description: String? = null,
    @SerialName("activity_type_id") val activityTypeId: String? = null,
    @SerialName("target_count") val targetCount: Int,
    @SerialName("start_date") val startDate: String? = null,
    @SerialName("end_date") val endDate: String? = null,
    @SerialName("created_at") val createdAt: String,
    val creator: UserDto? = null,
    @SerialName("activity_type") val activityType: ActivityTypeDto? = null,
    val participants: List<ChallengeParticipantDto>? = null
)

@Serializable
data class StreakDto(
    val id: String,
    @SerialName("group_id") val groupId: String,
    @SerialName("user_id") val userId: String,
    @SerialName("current_streak") val currentStreak: Int,
    @SerialName("longest_streak") val longestStreak: Int,
    @SerialName("last_activity_date") val lastActivityDate: String? = null,
    @SerialName("updated_at") val updatedAt: String,
    val user: UserDto? = null
)

@Serializable
data class GoalDto(
    val id: String,
    @SerialName("group_id") val groupId: String,
    @SerialName("user_id") val userId: String,
    @SerialName("activity_type_id") val activityTypeId: String? = null,
    @SerialName("target_count") val targetCount: Int,
    val period: String,
    @SerialName("current_progress") val currentProgress: Int,
    @SerialName("period_start") val periodStart: String,
    @SerialName("created_at") val createdAt: String,
    @SerialName("updated_at") val updatedAt: String,
    @SerialName("activity_type") val activityType: ActivityTypeDto? = null
)

@Serializable
data class CreateGoalRequest(
    @SerialName("activity_type_id") val activityTypeId: String? = null,
    @SerialName("target_count") val targetCount: Int,
    val period: String
)

@Serializable
data class UpdateGoalRequest(
    @SerialName("target_count") val targetCount: Int? = null,
    val period: String? = null
)

@Serializable
data class CreateChallengeRequest(
    val title: String,
    val description: String? = null,
    @SerialName("activity_type_id") val activityTypeId: String? = null,
    @SerialName("target_count") val targetCount: Int,
    @SerialName("start_date") val startDate: String? = null,
    @SerialName("end_date") val endDate: String? = null
)

@Serializable
data class BetaInviteDto(
    val id: String,
    val code: String,
    @SerialName("created_by") val createdBy: String,
    @SerialName("expires_at") val expiresAt: String,
    @SerialName("max_uses") val maxUses: Int,
    @SerialName("use_count") val useCount: Int,
    val note: String? = null,
    @SerialName("created_at") val createdAt: String,
    val creator: UserDto? = null
)

@Serializable
data class CreateBetaInviteRequest(
    @SerialName("max_uses") val maxUses: Int = 1,
    @SerialName("expires_in") val expiresIn: Int = 168,
    val note: String? = null
)
