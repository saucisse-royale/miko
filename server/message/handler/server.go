package handler

import (
	"errors"
	"git.emersion.fr/saucisse-royale/miko.git/server/message"
	"git.emersion.fr/saucisse-royale/miko.git/server/message/builder"
	"log"
)

var serverHandlers = &map[message.Type]TypeHandler{
	message.Types["version"]: func(ctx *message.Context, io *message.IO) error {
		var version message.ProtocolVersion
		read(io, &version)
		io.Version = version

		if version != message.CurrentVersion {
			code := message.ExitCodes["client_outdated"]
			if version > message.CurrentVersion {
				code = message.ExitCodes["server_outdated"]
			}

			if err := builder.SendExit(io, code); err != nil {
				return err
			}

			if err := io.Close(); err != nil {
				return err
			}
		} else {
			if err := builder.SendConfig(io, ctx.Config); err != nil {
				return err
			}
		}

		return nil
	},
	message.Types["exit"]: func(ctx *message.Context, io *message.IO) error {
		code := ReadExit(io)
		log.Println("Client disconnected with code:", code)

		if err := io.Close(); err != nil {
			return err
		}

		return nil
	},
	message.Types["login"]: func(ctx *message.Context, io *message.IO) error {
		username := readString(io)
		password := readString(io)

		code := ctx.Auth.Login(io.Id, username, password)
		if err := builder.SendLoginResp(io, code, ctx.Clock.GetRelativeTick()); err != nil {
			return err
		}

		log.Println("Client logged in:", username, code)

		if code != message.LoginResponseCodes["ok"] {
			return nil
		}

		// Create entity
		session := ctx.Auth.GetSession(io.Id)
		if session == nil {
			return errors.New("Cannot get newly logged in user's session")
		}

		// TODO: health default value
		session.Entity.Position.BX = 20
		session.Entity.Position.BY = 20
		session.Entity.Type = 0                                            // player
		session.Entity.Sprite = 1                                          // player
		session.Entity.Attributes[message.EntityAttrId(1)] = uint16(1000)  // health
		session.Entity.Attributes[message.EntityAttrId(30000)] = uint16(0) // cooldown_one

		req := ctx.Entity.Add(session.Entity, ctx.Clock.GetAbsoluteTick()) // TODO: move this elsewhere?
		err := req.Wait()
		log.Println("Entity registered!")
		if err != nil {
			return err
		}
		log.Println("Flushing entities diff")
		// Broadcast new entity
		err = builder.SendEntitiesDiffToClients(io.Broadcaster(), ctx.Clock.GetRelativeTick(), ctx.Entity.Flush())
		if err != nil {
			return err
		}
		log.Println("Sending terrain")
		// Send initial terrain
		pos := session.Entity.Position
		radius := message.BlockCoord(8)
		blks := []*message.Block{}
		for i := pos.BX - radius; i <= pos.BX+radius; i++ {
			for j := pos.BY - radius; j <= pos.BY+radius; j++ {
				blk, err := ctx.Terrain.GetBlockAt(i, j)
				if err != nil {
					log.Println("Cannot send initial terrain to client", err)
					continue
				}

				blks = append(blks, blk)
			}
		}

		err = builder.SendChunksUpdate(io, ctx.Clock.GetRelativeTick(), blks)
		if err != nil {
			return err
		}

		return builder.SendPlayerJoined(io.Broadcaster(), ctx.Clock.GetRelativeTick(), session.Entity.Id, username)
	},
	message.Types["register"]: func(ctx *message.Context, io *message.IO) error {
		username := readString(io)
		password := readString(io)

		code := ctx.Auth.Register(io.Id, username, password)

		log.Println("Client registered:", username, code)

		return builder.SendRegisterResp(io, code)
	},
	message.Types["terrain_request"]: func(ctx *message.Context, io *message.IO) error {
		var size uint8
		read(io, &size)

		for i := 0; i < int(size); i++ {
			var x, y message.BlockCoord
			read(io, &x)
			read(io, &y)

			blk, err := ctx.Terrain.GetBlockAt(x, y)
			if err != nil {
				return err
			}

			err = builder.SendChunkUpdate(io, ctx.Clock.GetRelativeTick(), blk)
			if err != nil {
				return err
			}
		}

		return nil
	},
	message.Types["entity_update"]: func(ctx *message.Context, io *message.IO) error {
		// TODO: add update initiator as parameter to ctx.Entity.Update()

		t := ctx.Clock.ToAbsoluteTick(readTick(io))

		entity, diff := ReadEntity(io)
		ctx.Entity.Update(entity, diff, t)

		return nil
	},
	message.Types["action_do"]: func(ctx *message.Context, io *message.IO) error {
		if !ctx.Auth.HasSession(io.Id) {
			return errors.New("Cannot execute action: remote not logged in")
		}

		session := ctx.Auth.GetSession(io.Id)

		action := &message.Action{
			Initiator: session.Entity.Id,
		}

		t := ctx.Clock.ToAbsoluteTick(readTick(io))
		read(io, &action.Id)

		// TODO: move action params somewhere else
		if action.Id == 0 { // throw_ball
			var angle float32
			var tmpId message.EntityId
			read(io, &angle)
			read(io, &tmpId)
			log.Println("Received ball action:", angle, tmpId)

			action.Params = []interface{}{angle, tmpId}
		} else {
			log.Println("Client triggered unknown action:", action.Id)
		}

		ctx.Action.Execute(action, t)
		return nil
	},
	message.Types["chat_send"]: func(ctx *message.Context, io *message.IO) error {
		msg := readString(io)

		if !ctx.Auth.HasSession(io.Id) {
			return errors.New("User not authenticated")
		}

		session := ctx.Auth.GetSession(io.Id)
		username := session.Username

		log.Println("Broadcasting chat message:", username, msg)

		return builder.SendChatReceive(io.Broadcaster(), ctx.Clock.GetRelativeTick(), username, msg)
	},
}
