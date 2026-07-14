package models

import (
	"context"
	"time"

	"github.com/mynaparrot/plugnmeet-protocol/plugnmeet"
	natsservice "github.com/mynaparrot/plugnmeet-server/pkg/services/nats"
	"github.com/mynaparrot/plugnmeet-server/pkg/dbmodels"
	"github.com/sirupsen/logrus"
)

func (m *JanitorModel) checkRoomWithDuration() {
	rooms := m.rm.GetRoomsWithDurationMap()
	for i, r := range rooms {
		now := uint64(time.Now().Unix())
		valid := r.StartedAt + (r.Duration * 60)
		if now > valid {
			_, _, _ = m.rm.EndRoom(context.Background(), &plugnmeet.RoomEndReq{
				RoomId: i,
			})
		}
	}
}

// activeRoomChecker will check & do reconciliation between DB & livekit
func (m *JanitorModel) activeRoomChecker() {
	log := m.logger.WithField("task", "activeRoomChecker")

	activeRooms, err := m.ds.GetActiveRoomsInfo(m.ctx)
	if err != nil {
		return
	}

	if len(activeRooms) == 0 {
		return
	}

	for _, room := range activeRooms {
		if room.Sid == "" {
			// if room RoomSid is empty then we won't do anything
			// because may be the session is creating
			// if we don't consider this, then it will unnecessarily create empty field
			continue
		}

		rInfo, err := m.natsService.GetRoomInfo(room.RoomId)
		if err != nil {
			log.WithError(err).Errorln("error getting room info")
			continue
		}

		// we did not find the room,
		// so, we're closing it
		if rInfo == nil {
			_, err = m.ds.UpdateRoomStatus(&dbmodels.RoomInfo{
				ID:        room.ID,
				IsRunning: 0,
			})
			if err != nil {
				log.WithError(err).Errorln("error updating room status")
			}
			continue
		}

		userIds, err := m.natsService.GetOnlineUsersId(room.RoomId)
		if err != nil {
			log.WithError(err).Errorln("error getting online users")
			continue
		}

		onlineCount := 0
		if userIds != nil {
			onlineCount = len(userIds)
		}

		if onlineCount > 0 {
			// Room is occupied — refresh last-occupied watermark.
			m.natsService.TouchRoomLastOccupiedAt(room.RoomId)
		} else {
			// Also treat "disconnected" (reconnect grace) as still occupied so a flaky
			// host/network blip does not start the empty timer prematurely.
			reconnectCount, err := m.natsService.CountUsersWithStatus(
				room.RoomId,
				natsservice.UserStatusDisconnected,
			)
			if err != nil {
				log.WithError(err).Warnln("error counting disconnected users")
			}
			if reconnectCount > 0 {
				m.natsService.TouchRoomLastOccupiedAt(room.RoomId)
				continue
			}

			lastOccupied := m.natsService.GetRoomLastOccupiedAt(room.RoomId)
			if lastOccupied == 0 {
				lastOccupied = rInfo.CreatedAt
			}
			valid := lastOccupied + rInfo.EmptyTimeout
			now := uint64(time.Now().UTC().Unix())
			if now > valid {
				log.WithFields(logrus.Fields{
					"emptyTimeout":  rInfo.EmptyTimeout,
					"createdAt":     rInfo.CreatedAt,
					"lastOccupied":  lastOccupied,
					"validUntil":    valid,
					"secondsOver":   now - valid,
				}).Info("closing empty room as it reached empty timeout since last occupancy")

				m.rm.EndRoom(context.Background(), &plugnmeet.RoomEndReq{RoomId: room.RoomId})
				continue
			}
		}

		var count = int64(onlineCount)
		if room.JoinedParticipants != count {
			_, _ = m.ds.UpdateNumParticipants(room.Sid, count)
		}
	}
}
