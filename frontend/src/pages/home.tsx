import { useContext, useEffect } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { GetMessages } from "../api/message";
import { GetPinnedMessages } from "../api/pinned_msgs";
import NavBar from "../components/NavBar";
import Routes from "../config";
import { ChannelContext, ChannelsContext } from "../contexts/channelctx";
import { ContextMenu } from "../contexts/context_menu_ctx";
import { PopUpContext } from "../contexts/popup";
import { UserContext, UserContextOBJ } from "../contexts/usercontext";
import { ChannelOBJ, MessageOBJ, ReadyOBJ, Status } from "../models/models";
import { Relationship as RelationshipOBJ } from "../models/relationship";
import PopUp from "./popup";
import { log } from "console";

function Home() {
	const user_ctx: UserContextOBJ = useContext(UserContext);
	const channel_context: ChannelContext = useContext(ChannelsContext);
	const popup_ctx = useContext(PopUpContext);
	const navigate = useNavigate();

	const SendNotification = (message: MessageOBJ) => {
		let url = location.pathname;
		if (!url.includes(message.channel_id)) {
			// TODO: если тот, кто отправил сообщение резко вышел из чата
			// то ему тоже придет Notification, поэтому нужно добавить еще фильтр на автора сообщения
			if (Notification.permission === "granted") {
				const notification = new Notification(message.author.username, {
					body: message.content,
					icon: message.author.avatar,
				});
				notification.onclick = () => {
					navigate(`/channels/${message.channel_id}`);
				};
			}
		}
	}

	useEffect(() => {
		if (Notification.permission === "default") {
			// TODO: не запрашивает
			Notification.requestPermission();
		}
		let last_ping = new Date().getTime();
		let last_pong = new Date().getTime();

		const NewGateway = () => {
			const gateway = new WebSocket(Routes.ws);
			last_ping = new Date().getTime();
			last_pong = new Date().getTime();
			return gateway;
		}

		let gateway = NewGateway();

		let interval = setInterval(() => {
			// если отправили пинг, а понг не пришел за 5 сек, то проблема....
			if (last_ping > last_pong || gateway.readyState > WebSocket.OPEN) { // CLOSING, CLOSED - нечитаемое условие
				console.log("Reconnecting...")
				gateway.close();
				location.reload();
			} else {
				const message = JSON.stringify({ event: "PING", data: "" });
				last_ping = new Date().getTime();
				gateway.send(message);
			}
		}, 5_000);

		// получение данных от сервера
		gateway.onmessage = (message: MessageEvent) => {
			const data = message.data;

			if (typeof data !== "string") {
				return
			}
			const payload = JSON.parse(data);

			switch (payload.event) {
				case "READY":
					// а что может в READY сообщении передать и все сообщения? или не стоит?
					console.log("connected to ws")
					const ready: ReadyOBJ = payload.data;
					user_ctx.setId(ready.user.id);
					user_ctx.setUsername(ready.user.username);
					user_ctx.setAvatar(ready.user.avatar);
					ready.channels.forEach((channel: ChannelOBJ) => {
						channel_context.setChannel(prev => new Map(prev.set(channel.id, channel)));
						// const userReactions = new Map(user_ctx.reactions);
						GetMessages(channel.id).then((msgs: MessageOBJ[]) => {
							channel_context.SetMessages(channel.id, msgs.reverse())
							// msgs.forEach(msg => {
							// 	msg.reactions?.forEach(reaction => {
							// 		if (reaction.user_id === ready.user.id) {
							// 			userReactions.set(msg.id, reaction.reaction);
							// 		}
							// 	});
							// });
							// user_ctx.setReactions(userReactions);
						})

						GetPinnedMessages(channel.id).then((msgs: MessageOBJ[]) => {
							channel_context.SetPinnedMessages(channel.id, msgs.reverse())
						})
					});
					console.log("finally reactions user", user_ctx.reactions)
					break;

				case "INVAILD_SESSION":
					// кука истекла?
					// TODO: сервер кажется такого не отправляет...
					navigate("/auth/login");
					break;

				case "PONG":
					last_pong = new Date().getTime();
					break;


				// т.е. сначала отправляем изменение на сервер, а тот по ws всем отправит, включая нам

				case "MESSAGE_CREATE":
					const message: MessageOBJ = payload.data;
					channel_context.InsertMessage(message);
					SendNotification(message);
					break;

				case "MESSAGE_MODIFY":
					const edited_message: MessageOBJ = payload.data;
					channel_context.UpdateMessage(edited_message);
					break;

				case "MESSAGE_DELETE":
					const deleted_message: MessageOBJ = payload.data;
					channel_context.DeleteMessage(deleted_message);
					break;

				case "CHANNEL_CREATE":
					const channel: ChannelOBJ = payload.data;
					channel_context.setChannel(prevChannels => new Map(prevChannels.set(channel.id, channel)));
					break;

				case "CHANNEL_MODIFY":
					const edited_channel: ChannelOBJ = payload.data;
					channel_context.setChannel(prevChannels => new Map(prevChannels.set(edited_channel.id, edited_channel)));
					break

				case "CHANNEL_DELETE":
					const deleted_channel: ChannelOBJ = payload.data;
					channel_context.deleteChannel(deleted_channel.id)
					break;

				case "MESSAGE_PINNED":
					const pinned_message: MessageOBJ = payload.data;
					channel_context.InsertPinnedMessage(pinned_message);
					break;

				case "MESSAGE_UNPINNED":
					const upinned_message: MessageOBJ = payload.data;
					channel_context.DeletePinnedMessage(upinned_message);
					break;

				case "RELATIONSHIP_CREATE":
					const new_relationship: RelationshipOBJ = payload.data;
					user_ctx.setRelationships(prevRelationship => new Map(prevRelationship.set(new_relationship.id, new_relationship)));
					break;

				case "RELATIONSHIP_MODIFY":
					const relationship: RelationshipOBJ = payload.data;
					user_ctx.setRelationships(prevRelationship => new Map(prevRelationship.set(relationship.id, relationship)));
					break;

				case "RELATIONSHIP_DELETE":
					const relationship_: RelationshipOBJ = payload.data;
					user_ctx.deleterelationship(relationship_.id);
					break;
			}

			// чел офлайн зашел в сеть
			if (payload.event === 'STATUS_UPDATE') {
				const status: Status = payload.data;

				// обновление статуса (онлайн -> офлайн | офлайн -> онлайн )
				// user_ctx
				if (status.type === 0) {
					const relationship_ = user_ctx.relationships.get(status.user_id);
					if (relationship_ !== undefined) {
						relationship_.status = status.status;
						user_ctx.setRelationships(prevFriends => new Map(prevFriends.set(relationship_.id, relationship_)));
					}
				}

				// обновление статусов участников каналов
				// channel_context
				if (status.type === 1) {
					const UpdateChannelStatus = (prevChannels: Map<String, ChannelOBJ>, status: Status) => {
						const channel = prevChannels.get(status.channel_id)
						if (channel === undefined) {
							return prevChannels
						}

						for (let i = 0; i < channel.recipients.length; i++) {
							if (channel.recipients[i].id === status.user_id) {
								channel.recipients[i].status = status.status;
								break;
							}
						}

						return prevChannels.set(channel.id, channel)

					}
					channel_context.setChannel(prevChannels => new Map(UpdateChannelStatus(prevChannels, status)));
				}
			}
		}

		// Home всегда примонтирован
		// когда вызовется функция?... при размонтировке?...
		return () => {
			console.log("cleaning up...");
			clearInterval(interval);
			gateway.close();
		};
	}, []);

	const ctx_menu = useContext(ContextMenu);



	useEffect(() => {
		const handleClick = () => {
			// каждый раз при клике по странице закрывается контекстное меню
			ctx_menu.close();
		};
		console.log('context menu use effect')
		window.addEventListener('click', handleClick);
		return () => window.removeEventListener('click', handleClick);
	}, []);


	return (
		<div className="h-screen w-full flex flex-col-reverse md:flex-row">
			<NavBar />
			<Outlet />
			{popup_ctx.show && <PopUp />}
			{ctx_menu.showCtxMenu && ctx_menu.ctxMenu}
		</div>
	);
}

export default Home;
