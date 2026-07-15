package models

import "github.com/mynaparrot/plugnmeet-protocol/plugnmeet"

// applyExplicitCreateRoomFeatureBools re-applies nested feature bools from the
// original create-room request after utils.PrepareDefaultRoomFeatures.
//
// Proto3 proto.Merge does not overwrite a non-zero default with a zero value
// (false). ZenLeader's MeetGatewayClient intentionally sends isAllow:false for
// whiteboard, recording, chat file upload, etc. — without this step those
// disables are silently dropped and the PlugNMeet defaults (true) win.
func applyExplicitCreateRoomFeatureBools(
	user *plugnmeet.RoomCreateFeatures,
	dest *plugnmeet.RoomCreateFeatures,
) {
	if user == nil || dest == nil {
		return
	}

	if user.RecordingFeatures != nil && dest.RecordingFeatures != nil {
		dest.RecordingFeatures.IsAllow = user.RecordingFeatures.IsAllow
		dest.RecordingFeatures.IsAllowCloud = user.RecordingFeatures.IsAllowCloud
		dest.RecordingFeatures.IsAllowLocal = user.RecordingFeatures.IsAllowLocal
		dest.RecordingFeatures.EnableAutoCloudRecording =
			user.RecordingFeatures.EnableAutoCloudRecording
		dest.RecordingFeatures.OnlyRecordAdminWebcams =
			user.RecordingFeatures.OnlyRecordAdminWebcams
	}

	if user.ChatFeatures != nil && dest.ChatFeatures != nil {
		dest.ChatFeatures.IsAllow = user.ChatFeatures.IsAllow
		dest.ChatFeatures.IsAllowFileUpload = user.ChatFeatures.IsAllowFileUpload
	}

	if user.WhiteboardFeatures != nil && dest.WhiteboardFeatures != nil {
		dest.WhiteboardFeatures.IsAllow = user.WhiteboardFeatures.IsAllow
		dest.WhiteboardFeatures.Visible = user.WhiteboardFeatures.Visible
	}

	if user.SharedNotePadFeatures != nil && dest.SharedNotePadFeatures != nil {
		dest.SharedNotePadFeatures.IsAllow = user.SharedNotePadFeatures.IsAllow
	}

	if user.DisplayExternalLinkFeatures != nil && dest.DisplayExternalLinkFeatures != nil {
		dest.DisplayExternalLinkFeatures.IsAllow =
			user.DisplayExternalLinkFeatures.IsAllow
	}

	if user.IngressFeatures != nil && dest.IngressFeatures != nil {
		dest.IngressFeatures.IsAllow = user.IngressFeatures.IsAllow
	}

	if user.ExternalBroadcastingFeatures != nil && dest.ExternalBroadcastingFeatures != nil {
		dest.ExternalBroadcastingFeatures.IsAllow =
			user.ExternalBroadcastingFeatures.IsAllow
		dest.ExternalBroadcastingFeatures.IsAllowRtmp =
			user.ExternalBroadcastingFeatures.IsAllowRtmp
	}

	if user.EndToEndEncryptionFeatures != nil && dest.EndToEndEncryptionFeatures != nil {
		dest.EndToEndEncryptionFeatures.IsEnabled =
			user.EndToEndEncryptionFeatures.IsEnabled
		dest.EndToEndEncryptionFeatures.IncludedChatMessages =
			user.EndToEndEncryptionFeatures.IncludedChatMessages
		dest.EndToEndEncryptionFeatures.IncludedWhiteboard =
			user.EndToEndEncryptionFeatures.IncludedWhiteboard
		dest.EndToEndEncryptionFeatures.EnabledSelfInsertEncryptionKey =
			user.EndToEndEncryptionFeatures.EnabledSelfInsertEncryptionKey
	}

	if user.InsightsFeatures != nil && dest.InsightsFeatures != nil {
		applyExplicitInsightsFeatureBools(user.InsightsFeatures, dest.InsightsFeatures)
	}
}

// applyExplicitInsightsFeatureBools copies nested Insights isAllow flags from the
// create request after proto.Merge (same proto3 zero-value issue).
func applyExplicitInsightsFeatureBools(
	user *plugnmeet.InsightsFeatures,
	dest *plugnmeet.InsightsFeatures,
) {
	if user == nil || dest == nil {
		return
	}
	dest.IsAllow = user.IsAllow
	if user.ChatTranslationFeatures != nil && dest.ChatTranslationFeatures != nil {
		dest.ChatTranslationFeatures.IsAllow = user.ChatTranslationFeatures.IsAllow
	}
	if user.AiFeatures != nil && dest.AiFeatures != nil {
		dest.AiFeatures.IsAllow = user.AiFeatures.IsAllow
		if user.AiFeatures.AiTextChatFeatures != nil &&
			dest.AiFeatures.AiTextChatFeatures != nil {
			dest.AiFeatures.AiTextChatFeatures.IsAllow =
				user.AiFeatures.AiTextChatFeatures.IsAllow
			dest.AiFeatures.AiTextChatFeatures.IsEnabled =
				user.AiFeatures.AiTextChatFeatures.IsEnabled
		}
		if user.AiFeatures.MeetingSummarizationFeatures != nil &&
			dest.AiFeatures.MeetingSummarizationFeatures != nil {
			dest.AiFeatures.MeetingSummarizationFeatures.IsAllow =
				user.AiFeatures.MeetingSummarizationFeatures.IsAllow
		}
	}
}

// enforceZenLeaderCreateRoomPolicy applies product-level disables that must not
// depend on callers remembering every flag (whiteboard is unsupported on mobile).
func enforceZenLeaderCreateRoomPolicy(rf *plugnmeet.RoomCreateFeatures) {
	if rf == nil {
		return
	}
	if rf.WhiteboardFeatures == nil {
		rf.WhiteboardFeatures = &plugnmeet.WhiteboardFeatures{}
	}
	rf.WhiteboardFeatures.IsAllow = false
	rf.WhiteboardFeatures.Visible = false
}
