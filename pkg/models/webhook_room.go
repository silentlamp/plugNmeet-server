package models

import (
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
	natsservice "github.com/mynaparrot/plugnmeet-server/pkg/services/nats"
	"github.com/sirupsen/logrus"
)

func (m *WebhookModel) roomStarted(event *livekit.WebhookEvent) {
	if event.Room == nil {
		m.logger.Warnln("received room_started webhook with nil room info")
		return
	}

	log := m.logger.WithFields(logrus.Fields{
		"roomId": event.Room.Name,
		"event":  event.GetEvent(),
	})
	log.Infoln("Handling room_started webhook")

	// we'll check the room from kv
	rInfo, meta, err := m.natsService.GetRoomInfoWithMetadata(event.Room.Name)
	if err != nil {
		log.WithError(err).Errorln("failed to get room info from NATS")
		return
	}

	if rInfo == nil || meta == nil {
		// Room was created directly in LiveKit (connection tester, ops probe, etc.)
		// without going through plugNmeet's API. Do NOT force-terminate — that
		// breaks LiveKit connection tests and any legitimate direct LiveKit use.
		// Orphan rooms still expire via LiveKit empty_timeout / room cleanup.
		log.Warnln("room not found in plugNmeet's NATS store, ignoring room_started webhook")
		return
	}

	if rInfo.Status != natsservice.RoomStatusActive {
		log.WithField("current_status", rInfo.Status).Info("updating room status to active")
		if err := m.natsService.UpdateRoomStatus(rInfo.RoomId, natsservice.RoomStatusActive); err != nil {
			log.WithError(err).Errorln("failed to update room status")
			return
		}
	}

	meta.StartedAt = uint64(time.Now().UTC().Unix())
	if meta.RoomFeatures.GetRoomDuration() > 0 {
		log.WithField("duration", meta.RoomFeatures.GetRoomDuration()).Info("adding room to duration checker")
		// we'll add room info in map
		err := m.rm.AddRoomWithDurationInfo(rInfo.RoomId, &RoomDurationInfo{
			Duration:  meta.RoomFeatures.GetRoomDuration(),
			StartedAt: meta.StartedAt,
		})
		if err != nil {
			log.WithError(err).Errorln("failed to add room duration info")
		}
	}

	if meta.IsBreakoutRoom {
		if err := m.bm.PostTaskAfterRoomStartWebhook(rInfo.RoomId, meta); err != nil {
			log.WithError(err).Errorln("failed to run post-start task for breakout room")
		}
	}

	if err := m.natsService.UpdateAndBroadcastRoomMetadata(rInfo.RoomId, meta); err != nil {
		log.WithError(err).Errorln("failed to update and broadcast room metadata")
	}

	// for room_started event we should send webhook at the end
	// otherwise some services may not be ready
	event.Room.Metadata = rInfo.Metadata
	event.Room.Sid = rInfo.RoomSid
	event.Room.MaxParticipants = uint32(rInfo.MaxParticipants)
	event.Room.EmptyTimeout = uint32(rInfo.EmptyTimeout)

	// webhook notification
	m.sendToWebhookNotifier(event)
	log.Info("Successfully processed room_started webhook")
}

func (m *WebhookModel) roomFinished(event *livekit.WebhookEvent) {
	if event.Room == nil {
		m.logger.Warnln("received room_finished webhook with nil room info")
		return
	}

	log := m.logger.WithFields(logrus.Fields{
		"roomId": event.Room.Name,
		"event":  event.GetEvent(),
	})
	log.Infoln("handling room_finished webhook")

	// Use the new helper function to get room info
	rInfo, err := m.getRoomInfoFromNatsOrRedis(event.Room.Name, log)
	if err != nil {
		log.WithError(err).Errorln("failed to get room info, skipping room_finished tasks")
		return
	}

	event.Room.Metadata = rInfo.Metadata
	event.Room.Sid = rInfo.RoomSid
	event.Room.MaxParticipants = uint32(rInfo.MaxParticipants)
	event.Room.EmptyTimeout = uint32(rInfo.EmptyTimeout)

	if rInfo.Status != natsservice.RoomStatusEnded {
		// LiveKit may emit room_finished when the media room briefly empties (host
		// reconnect, network blip). Do NOT end the PlugNMeet session if NATS still
		// has online or reconnecting users — otherwise everyone gets SESSION_ENDED.
		onlineIds, onlineErr := m.natsService.GetOnlineUsersId(rInfo.RoomId)
		reconnectCount, reconnectErr := m.natsService.CountUsersWithStatus(
			rInfo.RoomId,
			natsservice.UserStatusDisconnected,
		)
		stillOccupied := (onlineErr == nil && len(onlineIds) > 0) ||
			(reconnectErr == nil && reconnectCount > 0)
		if stillOccupied {
			log.WithFields(logrus.Fields{
				"online":       len(onlineIds),
				"disconnected": reconnectCount,
			}).Warnln("ignoring LiveKit room_finished; NATS still has occupants/reconnecters")
			return
		}

		// This means the room was ended directly by LiveKit (e.g., empty timeout),
		// not through the plugNmeet API. We need to trigger our cleanup flow.
		log.Warnln("room was not ended via API, triggering plugNmeet EndRoom flow")

		// change status to ended
		if err := m.natsService.UpdateRoomStatus(rInfo.RoomId, natsservice.RoomStatusEnded); err != nil {
			log.WithError(err).Errorln("failed to update room status")
		}
		// end the room in the proper plugNmeet way
		m.rm.EndRoom(m.ctx, &plugnmeet.RoomEndReq{RoomId: rInfo.RoomId})
	}

	// at the end we'll handle event notification
	m.sendToWebhookNotifier(event)

	log.Info("Successfully processed room_finished webhook")
	// webhook data will be clean after analytics export method call e.g. PrepareToExportAnalytics
}
