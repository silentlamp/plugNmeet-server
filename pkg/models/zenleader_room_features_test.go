package models

import (
	"testing"

	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
	"github.com/mynaparrot/plugnmeet-protocol/utils"
	"google.golang.org/protobuf/proto"
)

// TestProto3FalseBoolsHonoredAfterPrepare reproduces the PlugNMeet Merge bug:
// sending isAllow:false for nested features must stick after PrepareDefaultRoomFeatures.
func TestProto3FalseBoolsHonoredAfterPrepare(t *testing.T) {
	req := &plugnmeet.CreateRoomReq{
		RoomId: "test-room",
		Metadata: &plugnmeet.RoomMetadata{
			RoomTitle: "Test",
			RoomFeatures: &plugnmeet.RoomCreateFeatures{
				WhiteboardFeatures: &plugnmeet.WhiteboardFeatures{IsAllow: false},
				ChatFeatures: &plugnmeet.ChatFeatures{
					IsAllow:           true,
					IsAllowFileUpload: false,
				},
				RecordingFeatures: &plugnmeet.RecordingFeatures{
					IsAllow:      false,
					IsAllowCloud: false,
					IsAllowLocal: false,
				},
			},
		},
	}

	userRf := proto.Clone(req.Metadata.RoomFeatures).(*plugnmeet.RoomCreateFeatures)
	utils.PrepareDefaultRoomFeatures(req)

	// Without override, whiteboard stays true (proto3 Merge).
	if !req.Metadata.RoomFeatures.WhiteboardFeatures.IsAllow {
		t.Fatal("expected upstream PrepareDefault to keep whiteboard default true before override")
	}

	applyExplicitCreateRoomFeatureBools(userRf, req.Metadata.RoomFeatures)
	enforceZenLeaderCreateRoomPolicy(req.Metadata.RoomFeatures)

	rf := req.Metadata.RoomFeatures
	if rf.WhiteboardFeatures.IsAllow {
		t.Error("whiteboard should be disabled after ZenLeader create policy")
	}
	if rf.WhiteboardFeatures.Visible {
		t.Error("whiteboard visible should be false")
	}
	if rf.ChatFeatures.IsAllowFileUpload {
		t.Error("chat file upload should stay false from create request")
	}
	if rf.RecordingFeatures.IsAllow {
		t.Error("recording should stay false from create request")
	}
}
