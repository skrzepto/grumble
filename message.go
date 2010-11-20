// Grumble - an implementation of Murmur in Go
// Copyright (c) 2010 The Grumble Authors
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package main

import (
	"log"
	"mumbleproto"
	"goprotobuf.googlecode.com/hg/proto"
	"net"
	"cryptstate"
)

// These are the different kinds of messages
// that are defined for the Mumble protocol
const (
	MessageVersion = iota
	MessageUDPTunnel
	MessageAuthenticate
	MessagePing
	MessageReject
	MessageServerSync
	MessageChannelRemove
	MessageChannelState
	MessageUserRemove
	MessageUserState
	MessageBanList
	MessageTextMessage
	MessagePermissionDenied
	MessageACL
	MessageQueryUsers
	MessageCryptSetup
	MessageContextActionAdd
	MessageContextAction
	MessageUserList
	MessageVoiceTarget
	MessagePermissionQuery
	MessageCodecVersion
	MessageUserStats
	MessageRequestBlob
	MessageServerConfig
)

const (
	UDPMessageVoiceCELTAlpha = iota
	UDPMessagePing
	UDPMessageVoiceSpeex
	UDPMessageVoiceCELTBeta
)

type Message struct {
	buf []byte

	// Kind denotes a message kind for TCP packets. This field
	// is ignored for UDP packets.
	kind uint16

	// For UDP datagrams one of these fields have to be filled out.
	// If there is no connection established, address must be used.
	// If the datagram comes from an already-connected client, the
	// client field should point to that client.
	client  *Client
	address net.Addr
}

type VoiceBroadcast struct {
	// The client who is performing the broadcast
	client *Client
	// The VoiceTarget identifier.
	target byte
	// The voice packet itself.
	buf []byte
}

func (server *Server) handleCryptSetup(client *Client, msg *Message) {
	cs := &mumbleproto.CryptSetup{}
	err := proto.Unmarshal(msg.buf, cs)
	if err != nil {
		client.Panic(err.String())
		return
	}

	// No client nonce. This means the client
	// is requesting that we re-sync our nonces.
	if len(cs.ClientNonce) == 0 {
		log.Printf("Requested crypt-nonce resync")
		cs.ClientNonce = make([]byte, cryptstate.AESBlockSize)
		if copy(cs.ClientNonce, client.crypt.EncryptIV[0:]) != cryptstate.AESBlockSize {
			return
		}
		client.sendProtoMessage(MessageCryptSetup, cs)
	} else {
		log.Printf("Received client nonce")
		if len(cs.ClientNonce) != cryptstate.AESBlockSize {
			return
		}

		client.crypt.Resync += 1
		if copy(client.crypt.DecryptIV[0:], cs.ClientNonce) != cryptstate.AESBlockSize {
			return
		}
		log.Printf("Crypt re-sync successful")
	}
}

func (server *Server) handlePingMessage(client *Client, msg *Message) {
	ping := &mumbleproto.Ping{}
	err := proto.Unmarshal(msg.buf, ping)
	if err != nil {
		client.Panic(err.String())
		return
	}

	// Phony response for ping messages. We don't keep stats
	// for this yet.
	client.sendProtoMessage(MessagePing, &mumbleproto.Ping{
		Timestamp: ping.Timestamp,
		Good:      proto.Uint32(uint32(client.crypt.Good)),
		Late:      proto.Uint32(uint32(client.crypt.Late)),
		Lost:      proto.Uint32(uint32(client.crypt.Lost)),
		Resync:    proto.Uint32(uint32(client.crypt.Resync)),
	})
}

func (server *Server) handleChannelAddMessage(client *Client, msg *Message) {
}

func (server *Server) handleChannelRemoveMessage(client *Client, msg *Message) {
}

func (server *Server) handleChannelStateMessage(client *Client, msg *Message) {
}

func (server *Server) handleUserRemoveMessage(client *Client, msg *Message) {
}

func (server *Server) handleUserStateMessage(client *Client, msg *Message) {
	log.Printf("UserState!")
	userstate := &mumbleproto.UserState{}
	err := proto.Unmarshal(msg.buf, userstate)
	if err != nil {
		client.Panic(err.String())
	}

	if userstate.Session == nil {
		log.Printf("UserState without session.")
		return
	}

	actor := server.clients[client.Session]
	user := server.clients[*userstate.Session]

	log.Printf("actor = %v", actor)
	log.Printf("user = %v", user)

	userstate.Session = proto.Uint32(user.Session)
	userstate.Actor = proto.Uint32(actor.Session)

	// Has a channel ID
	if userstate.ChannelId != nil {
		// Destination channel
		dstChan := server.channels[int(*userstate.ChannelId)]
		log.Printf("dstChan = %v", dstChan)

		// If the user and the actor aren't the same, check whether the actor has the 'move' permission
		// on the user's channel to move.

		// Check whether the actor has 'move' permissions on dstChan.  Check whether user has 'enter'
		// permissions on dstChan.

		// Check whether the channel is full.
	}

	if userstate.Mute != nil || userstate.Deaf != nil || userstate.Suppress != nil || userstate.PrioritySpeaker != nil {
		// Disallow for SuperUser

		// Check whether the actor has 'mutedeafen' permission on user's channel.

		// Check if this was a suppress operation. Only the server can suppress users.
	}

	// Comment set/clear
	if userstate.Comment != nil {
		comment := *userstate.Comment
		log.Printf("comment = %v", comment)

		// Clearing another user's comment.
		if user != actor {
			// Check if actor has 'move' permissions on the root channel. It is needed
			// to clear another user's comment.

			// Only allow empty text.
		}

		// Check if the text is allowed.

		// Only set the comment if it is different from the current
		// user comment.
	}

	// Texture change
	if userstate.Texture != nil {
		// Check the length of the texture
	}

	// Registration
	if userstate.UserId != nil {
		// If user == actor, check for 'selfregister' permission on root channel.
		// If user != actor, check for 'register' permission on root channel.

		// Check if the UserId in the message is >= 0. A registration attempt
		// must use a negative UserId.
	}

	// Prevent self-targetting state changes to be applied to other users
	// That is, if actor != user, then:
	//   Discard message if it has any of the following things set:
	//      - SelfDeaf
	//      - SelfMute
	//      - Texture
	//      - PluginContext
	//      - PluginIdentity
	//      - Recording
	if actor != user && (userstate.SelfDeaf != nil || userstate.SelfMute != nil ||
		userstate.Texture != nil || userstate.PluginContext != nil || userstate.PluginIdentity != nil ||
		userstate.Recording != nil) {
			return
	}

}

func (server *Server) handleBanListMessage(client *Client, msg *Message) {
}

func (server *Server) handleTextMessage(client *Client, msg *Message) {
	txtmsg := &mumbleproto.TextMessage{}
	err := proto.Unmarshal(msg.buf, txtmsg)
	if err != nil {
		client.Panic(err.String())
		return
	}

	users := []*Client{}
	for i := 0; i < len(txtmsg.Session); i++ {
		user, ok := server.clients[txtmsg.Session[i]]
		if !ok {
			log.Panic("Could not look up client by session")
		}
		users = append(users, user)
	}

	for _, user := range users {
		user.sendProtoMessage(MessageTextMessage, &mumbleproto.TextMessage{
			Actor:   proto.Uint32(client.Session),
			Message: txtmsg.Message,
		})
	}
}

func (server *Server) handleAclMessage(client *Client, msg *Message) {
}

// User query
func (server *Server) handleQueryUsers(client *Client, msg *Message) {
}

// User stats message. Shown in the Mumble client when a
// user right clicks a user and selects 'User Information'.
func (server *Server) handleUserStatsMessage(client *Client, msg *Message) {
	stats := &mumbleproto.UserStats{}
	err := proto.Unmarshal(msg.buf, stats)
	if err != nil {
		client.Panic(err.String())
	}
	log.Printf("UserStatsMessage")
}
