package models

import "github.com/mynaparrot/plugnmeet-protocol/plugnmeet"

// applyRequestedFalseOverrides re-applies proto3 false bools from the API create
// request that proto.Merge cannot overwrite onto default-true nested features.
//
// ZenLeader's Java MeetGatewayClient sends isAllow:false for whiteboard,
// recording, notepad, RTMP, ingress, external link, chat file upload, and AI.
// Without this step those capabilities stay enabled from protocol defaults.
func applyRequestedFalseOverrides(dst, src *plugnmeet.RoomCreateFeatures) {
	if dst == nil || src == nil {
		return
	}

	if src.AllowVirtualBg != nil && !*src.AllowVirtualBg {
		f := false
		dst.AllowVirtualBg = &f
	}

	if src.WhiteboardFeatures != nil && !src.WhiteboardFeatures.IsAllow && dst.WhiteboardFeatures != nil {
		dst.WhiteboardFeatures.IsAllow = false
	}

	if src.RecordingFeatures != nil && dst.RecordingFeatures != nil {
		if !src.RecordingFeatures.IsAllow {
			dst.RecordingFeatures.IsAllow = false
			dst.RecordingFeatures.IsAllowCloud = false
			dst.RecordingFeatures.IsAllowLocal = false
			dst.RecordingFeatures.EnableAutoCloudRecording = false
		} else {
			if !src.RecordingFeatures.IsAllowCloud {
				dst.RecordingFeatures.IsAllowCloud = false
			}
			if !src.RecordingFeatures.IsAllowLocal {
				dst.RecordingFeatures.IsAllowLocal = false
			}
			if !src.RecordingFeatures.EnableAutoCloudRecording {
				dst.RecordingFeatures.EnableAutoCloudRecording = false
			}
		}
	}

	if src.ChatFeatures != nil && dst.ChatFeatures != nil && !src.ChatFeatures.IsAllowFileUpload {
		dst.ChatFeatures.IsAllowFileUpload = false
	}

	if src.SharedNotePadFeatures != nil && !src.SharedNotePadFeatures.IsAllow && dst.SharedNotePadFeatures != nil {
		dst.SharedNotePadFeatures.IsAllow = false
	}

	if src.DisplayExternalLinkFeatures != nil && !src.DisplayExternalLinkFeatures.IsAllow && dst.DisplayExternalLinkFeatures != nil {
		dst.DisplayExternalLinkFeatures.IsAllow = false
	}

	if src.IngressFeatures != nil && !src.IngressFeatures.IsAllow && dst.IngressFeatures != nil {
		dst.IngressFeatures.IsAllow = false
	}

	if src.ExternalBroadcastingFeatures != nil && dst.ExternalBroadcastingFeatures != nil {
		if !src.ExternalBroadcastingFeatures.IsAllow {
			dst.ExternalBroadcastingFeatures.IsAllow = false
			dst.ExternalBroadcastingFeatures.IsAllowRtmp = false
		} else if !src.ExternalBroadcastingFeatures.IsAllowRtmp {
			dst.ExternalBroadcastingFeatures.IsAllowRtmp = false
		}
	}

	if src.InsightsFeatures == nil || dst.InsightsFeatures == nil {
		return
	}

	if src.InsightsFeatures.ChatTranslationFeatures != nil &&
		!src.InsightsFeatures.ChatTranslationFeatures.IsAllow &&
		dst.InsightsFeatures.ChatTranslationFeatures != nil {
		dst.InsightsFeatures.ChatTranslationFeatures.IsAllow = false
	}

	if src.InsightsFeatures.AiFeatures == nil || dst.InsightsFeatures.AiFeatures == nil {
		return
	}
	aiSrc := src.InsightsFeatures.AiFeatures
	aiDst := dst.InsightsFeatures.AiFeatures
	if !aiSrc.IsAllow {
		aiDst.IsAllow = false
	}
	if aiSrc.AiTextChatFeatures != nil && !aiSrc.AiTextChatFeatures.IsAllow && aiDst.AiTextChatFeatures != nil {
		aiDst.AiTextChatFeatures.IsAllow = false
		aiDst.AiTextChatFeatures.IsEnabled = false
	}
	if aiSrc.MeetingSummarizationFeatures != nil && !aiSrc.MeetingSummarizationFeatures.IsAllow && aiDst.MeetingSummarizationFeatures != nil {
		aiDst.MeetingSummarizationFeatures.IsAllow = false
	}
}
