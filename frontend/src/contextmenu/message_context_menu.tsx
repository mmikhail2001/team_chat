import React, { useContext, useEffect, useState } from 'react'
import { MessageOBJ } from '../models/models';
import { UserContext } from "../contexts/usercontext";
import { MessageContext } from "../contexts/messagectx";
import { ChannelsContext } from "../contexts/channelctx";
import { ThreadContext, ThreadContextOBJ } from "../contexts/threadcontext";
import Routes from '../config'
import DeleteMessage from '../components/popup/DeleteMessage';
import { PopUpContext } from '../contexts/popup';
import { FaHeart, FaSmile } from 'react-icons/fa';
import { BsFillHeartFill, BsFillXDiamondFill } from 'react-icons/bs';
import { BiSolidLike } from "react-icons/bi";
import { BiSolidDislike } from "react-icons/bi";


interface propsMsgCtxProps {
    x: number, y: number, message: MessageOBJ
}

export default function MessageContextMenu(props: propsMsgCtxProps) {
    const message = props.message;
    console.log('---->', message.content)
    const user_ctx = useContext(UserContext);
    const popup_ctx = useContext(PopUpContext);
    const channel_ctx = useContext(ChannelsContext);
    const msgctx = useContext(MessageContext);
    const thread_context: ThreadContextOBJ = useContext(ThreadContext);

    const channel = channel_ctx.channels.get(props.message.channel_id);
    if (!channel) {
        throw new Error("key on map not exists (channel)");
    }

    const [isPinned, setIsPinned] = useState(false);
    useEffect(() => {
        const pinnedMessage = channel_ctx.pinnedMessages.get(message.channel_id);
        if (pinnedMessage !== undefined) {
            for (let i = 0; i < pinnedMessage.length; i++) {
                const messageFound = pinnedMessage[i].id === message.id
                setIsPinned(messageFound ? true : false)
                if (messageFound) break
            }
        } else {
            setIsPinned(false);
        }
        console.log("isPinned ===", isPinned)
    }, [channel_ctx.pinnedMessages, message.id])
    // если держать контекстное меню включенным и другой пользователь сделает pin
    // то в режиме реального времени в данном контекстном меню все изменится
    // pin -> unpin (однако наоборот нет....)



    let style: React.CSSProperties
    style = {
        top: props.y,
        left: props.x
    }

    function PinMsg() {
        const url = Routes.Channels + '/' + message.channel_id + '/pins/' + message.id;
        fetch(url, {
            method: 'PUT'
        }).then(res => {
            if (res.status === 200) {
                channel_ctx.UpdatePinnedMessage(message);
            }
        })
    }

    function UnpinMsg() {
        const url = Routes.Channels + '/' + message.channel_id + '/pins/' + message.id;
        fetch(url, {
            method: 'DELETE'
        }).then(res => {
            if (res.status === 200) {
                channel_ctx.DeletePinnedMessage(message);
            }
        })
    }

    const openThread = () => {
        if (!message.has_thread) {
            // Отправка POST запроса для создания нового канала
            const url = `${Routes.Channels}/${message.channel_id}/messages/${message.id}`;
            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            })
                .then(response => {
                    if (response.ok || response.status === 400) {
                        console.log("request openThread 200")
                        return response.json();
                    }
                    throw new Error('Failed to create new channel');
                })
                .then(newChannel => {
                    // Добавляем новый канал в контекст каналов
                    channel_ctx.setChannel(prevChannels => new Map(prevChannels).set(newChannel.id, newChannel));
                    // Обновляем поле thread_id сообщения
                    const notUpdatedMessages = channel_ctx.messages.get(message.channel_id)
                    if (!notUpdatedMessages) {
                        throw new Error("key on map not exists (notUpdatedMessages)");
                    }
                    const updatedMessages = notUpdatedMessages.map(messageMap => {
                        if (message.id === messageMap.id) {
                            console.log("update message", message.content)
                            return {
                                ...messageMap,
                                has_thread: true,
                                thread_id: newChannel.id,
                            };
                        }
                        return messageMap;
                    });
                    channel_ctx.SetMessages(message.channel_id, updatedMessages);
                    // Устанавливаем канал и сообщение в контексте треда и отображаем тред
                    thread_context.setChannel(channel);
                    thread_context.setThread(newChannel);
                    thread_context.setMessage(message);
                    thread_context.setThreadShow(true);
                })
                .catch(error => {
                    console.error('Error creating new channel:', error);
                });
        } else {
            const threadMessages = channel_ctx.channels.get(message.thread_id)
            if (!threadMessages) {
                throw new Error("key on map not exists (threadMessages)");
            }
            thread_context.setThread(threadMessages);
            thread_context.setChannel(channel);
            thread_context.setMessage(message);
            thread_context.setThreadShow(true);
        }
    };

    const sendReaction = (reaction: string) => {
        const url = `${Routes.Messages}/${message.id}/react?reaction=${reaction}`;
        fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        })
            .then(response => response.json())
            .then(updatedMessage => {
                channel_ctx.UpdateMessage(updatedMessage);
            })
            .catch(error => console.error('Error sending reaction:', error));
    };



    return (
        <div className='ContextMenu' style={style}>
            <div className="flex gap-2 justify-center">
                <FaHeart className='reaction-icon text-2xl text-red-500 cursor-pointer hover:text-red-600' onClick={() => sendReaction('love')} />
                <FaSmile className='reaction-icon text-2xl text-yellow-500 cursor-pointer hover:text-yellow-600' onClick={() => sendReaction('smile')} />
                <BiSolidLike className='reaction-icon text-2xl text-green-600 cursor-pointer hover:text-green-700' onClick={() => sendReaction('like')} />
                <BiSolidDislike className='reaction-icon text-2xl text-blue-500 cursor-pointer hover:text-blue-600' onClick={() => sendReaction('dislike')} />
            </div>
            <button className='CtxBtn' onClick={() => { navigator.clipboard.writeText(props.message.content) }}>Copy Text</button>
            {!isPinned && <button className='CtxBtn' onClick={PinMsg}>Pin Message</button>}
            {isPinned && <button className='CtxBtn' onClick={UnpinMsg}>Unpin Message</button>}
            {user_ctx.id === message.author.id && <button className='CtxBtn' onClick={() => { msgctx.setMessage(message); msgctx.setMessageEdit(true) }}>Edit Message</button>}
            {(user_ctx.id === message.author.id || channel?.owner_id === user_ctx.id) && <button className='CtxDelBtn' onClick={() => popup_ctx.open(<DeleteMessage message={message} />)}>Delete Message</button>}

            {/* наверное здесь нужно проверять, есть ли у данного сообщения thread_id, если нет, то отправлять запрос на создание */}
            {(channel?.type === 2 && !message.system_message) && <button className='CtxBtn'
                onClick={openThread}>Open thread</button>}

            <button className='CtxBtn' onClick={() => navigator.clipboard.writeText(props.message.id)}>Copy ID</button>
        </div>
    )
}
