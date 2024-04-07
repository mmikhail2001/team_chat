import { useState, useEffect, useContext, useRef } from "react";
import { setDefaultAvatar } from '../../utils/errorhandle';
import { MessageContext } from "../../contexts/messagectx";
import { MessageOBJ } from "../../models/models";
import { RiPencilFill, RiQuestionAnswerLine } from "react-icons/ri";
import { UserContext } from "../../contexts/usercontext";
import { ChannelContext, ChannelsContext } from "../../contexts/channelctx";
import Routes from "../../config";
import AttachmentDefault from "./attachment/default";
import AttachmentImage from "./attachment/image";
import AttachmentVideo from "./attachment/video";
import AttachmentAudio from "./attachment/audio";
import { FaServer } from "react-icons/fa";
import { ContextMenu } from "../../contexts/context_menu_ctx";
import MessageContextMenu from "../../contextmenu/message_context_menu";
import { MdOutlineQuestionAnswer } from "react-icons/md";
import { FaHeart, FaSmile } from 'react-icons/fa';
import { BiSolidLike } from "react-icons/bi";
import { BiSolidDislike } from "react-icons/bi";


function Message({ message, short }: { message: MessageOBJ, short: boolean }) {
    const msgctx = useContext(MessageContext);
    const user_ctx = useContext(UserContext);
    const channel_ctx = useContext(ChannelsContext);
    // const messageElement = useRef<HTMLDivElement>(null);
    const ctx_menu = useContext(ContextMenu);

    const [edit, setEdit] = useState(false);
    const [msg, setMsg] = useState(message.content);

    const [isBlocked, setIsBlocked] = useState(false);
    const [ShowMsg, setShowMsg] = useState(true);

    const [attachmentElement, setAttachmentElement] = useState<JSX.Element>(<></>);

    let time = new Date(message.created_at * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

    // const [reactions, setReactions] = useState<ReactionOBJ[]>([]);
    const [reactionElements, setReactionElements] = useState<JSX.Element[]>([]);


    // const [isInitialRender, setIsInitialRender] = useState(true);

    // useEffect(() => {
    //     setIsInitialRender(false);
    // }, []);

    useEffect(() => {
        if (message.attachments.length > 0) {
            const file = message.attachments[0]

            if (file.content_type.search(/image\/.+/) !== -1) {
                setAttachmentElement(<AttachmentImage message={message} />)
            } else if (file.content_type.search(/video\/.+/) !== -1) {
                setAttachmentElement(<AttachmentVideo message={message} />)
            } else if (file.content_type.search(/audio\/.+/) !== -1) {
                setAttachmentElement(<AttachmentAudio message={message} />)
            } else {
                setAttachmentElement(<AttachmentDefault message={message} />)
            }
        }
    }, [message]);

    useEffect(() => {
        if (msgctx.messageEdit && msgctx.message.id === message.id) {
            setEdit(true);
        } else {
            setEdit(false);
        }
    }, [msgctx.messageEdit, msgctx.message, message]);

    function handleEdit() {
        const url = Routes.Channels + "/" + message.channel_id + "/messages/" + message.id;
        fetch(url, {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ content: msg })
        })
        msgctx.setMessageEdit(false);
        setEdit(false);
    }

    function cancelEdit() {
        setMsg(message.content)
        setEdit(false)
        msgctx.setMessageEdit(false)
    }

    function handleKey(event: React.KeyboardEvent<HTMLInputElement>) {
        if (event.key === 'Enter') {
            handleEdit();
        }
        if (event.key === 'Escape') {
            cancelEdit();
        }
    }

    useEffect(() => {
        const author_id = message.author.id;
        const relationship = user_ctx.relationships.get(author_id);
        if (relationship) {
            if (relationship.type === 2) {
                setIsBlocked(true);
                setShowMsg(false)
            }
        }
    }, [user_ctx.relationships, message.author])

    function onInputChange(event: React.ChangeEvent<HTMLInputElement>) {
        const inputstr = event.target.value;
        if (inputstr.length <= 2000) {
            setMsg(inputstr);
        } else {
            alert("Message too long");
        }
    }

    const handleReactionClick = (reactionType: string) => {
        console.log("handleReactionClick");
    
        // Проверяем, была ли уже поставлена данная реакция пользователем на данном сообщении
        if (user_ctx.reactions.has(message.id) && user_ctx.reactions.get(message.id) === reactionType) {
            // Если реакция уже стоит, то нужно ее удалить
            fetch(`/api/messages/${message.id}/react`, {
                method: 'DELETE'
            })
            .then(response => {
                if (response.ok) {
                    response.json().then((updatedMessage: MessageOBJ) => {
                        // Обновляем сообщение в контексте канала
                        channel_ctx.UpdateMessage(updatedMessage);
                        // Обновляем реакции пользователя
                        const updatedReactions = new Map(user_ctx.reactions);
                        updatedReactions.delete(message.id);
                        user_ctx.setReactions(updatedReactions);
                    });
                }
            })
            .catch(error => console.error('Error removing reaction:', error));
        } else {
            // Устанавливаем новую реакцию
            fetch(`/api/messages/${message.id}/react?reaction=${reactionType}`, {
                method: 'POST'
            })
            .then(response => {
                if (response.ok) {
                    // Получаем обновленную структуру сообщения с сервера
                    response.json().then((updatedMessage: MessageOBJ) => {
                        // Обновляем сообщение в контексте канала
                        channel_ctx.UpdateMessage(updatedMessage);
                        // Обновляем реакции пользователя
                        const updatedReactions = new Map(user_ctx.reactions);
                        updatedReactions.set(message.id, reactionType);
                        user_ctx.setReactions(updatedReactions);
                    });
                }
            })
            .catch(error => console.error('Error setting reaction:', error));
        }
    };

    useEffect(() => {
        console.log("useEffect set react")
        const updateReactions = () => {
            const reactionCounts: { [key: string]: number } = {};
            console.log(">>>> message.reactions ", message.reactions)
            message.reactions?.forEach(reaction => {
                if (!reactionCounts[reaction.reaction]) {
                    reactionCounts[reaction.reaction] = 0;
                }
                reactionCounts[reaction.reaction]++;
            });

    
            const reactionsMap: { [key: string]: React.ReactNode } = {
                'love': <FaHeart className='reaction-icon text-2xl text-red-500 cursor-pointer hover:text-red-600'/>,
                'smile': <FaSmile className='reaction-icon text-2xl text-yellow-500 cursor-pointer hover:text-yellow-600'/>,
                'like': <BiSolidLike className='reaction-icon text-2xl text-green-600 cursor-pointer hover:text-green-700' />,
                'dislike': <BiSolidDislike className='reaction-icon text-2xl text-blue-500 cursor-pointer hover:text-blue-600'/>
            };
    
            const reactionElements: JSX.Element[] = [];
            Object.keys(reactionsMap).forEach(reactionType => {
                const count = reactionCounts[reactionType] || 0;
                if (count > 0) {
                    const isUserReaction = user_ctx.reactions.has(message.id) && user_ctx.reactions.get(message.id) === reactionType;
                    console.log(reactionType, message.id, isUserReaction)
                    console.log("user_ctx.reactions ===", user_ctx.reactions)
                    const reactionIcon = reactionsMap[reactionType];
                    const element = (
                        <div key={reactionType} className={`flex gap-2 ${isUserReaction ? 'bg-sky-600 p-1 rounded-lg px-2' : ''}`} onClick={() => handleReactionClick(reactionType)}>
                            <div className="flex items-center">
                                {reactionIcon}
                                {isUserReaction ? <span className="font-bold ml-1">{count}</span> : <span className="ml-1">{count}</span>}
                            </div>
                        </div>
                    );
                    reactionElements.push(element);
                }
            });
            return reactionElements;
        };
        setReactionElements(updateReactions());
    }, [message.reactions, user_ctx.reactions]);


    return (
        <div className="relative w-full flex my-1 hover:bg-zinc-900" onContextMenu={(event) => {
            event.preventDefault();
            ctx_menu.open(<MessageContextMenu x={event.clientX} y={event.clientY} message={message} />)
        }
        }>
            <div className="absolute left-0 w-24 flex items-center justify-center">
                {(!message.system_message && !short && ShowMsg) && <img className="h-12 w-12 rounded-xl bg-zinc-800" src={message.author.avatar} alt="Avatar" onError={setDefaultAvatar} />}
                {message.system_message && <FaServer size={24} />}
            </div>
            <div className="w-full ml-24 mr-32 flex flex-col">
                {ShowMsg && <>
                    {(!message.system_message && !short) && <span className="text-xl">{message.author.username}</span>}
                    {!edit && 
                        <div className="text-neutral-400">
                        <div className={user_ctx.id === message.author.id ? "bg-gray-800 p-2 rounded-lg inline-block" : "bg-gray-700 bg-opacity-50 p-2 rounded-lg inline-block"}>
                            {message.content}
                        </div>
                      </div>
                        // <span className="text-neutral-400 rounded-lg bg-gray-800 bg-opacity-50 p-2 inline-block">{message.content}</span>
                    }
                    {edit &&
                        <div>
                            <input className="bg-zinc-800 w-11/12 outline-none px-2 rounded" type="text" value={msg} onKeyDown={handleKey} onChange={onInputChange} />
                            <p className="text-xm">Escape to <button className="text-cyan-400 text-sm hover:underline" onClick={cancelEdit}>Cancel</button> • Enter to <button className="text-cyan-400 text-sm hover:underline" onClick={handleEdit}>Save</button></p>
                        </div>
                    }
                    {attachmentElement}
                    {isBlocked && <p className="text-cyan-500 text-xs cursor-pointer" onClick={() => { setShowMsg(false) }}>Hide</p>}
                </>}
                {!ShowMsg && <p>Message From User You Blocked! <button className="text-cyan-500" onClick={() => { setShowMsg(true) }}>Reveal</button></p>}
                <div className="flex gap-2 my-2">
                    {reactionElements}
                </div>
            </div>
            <div className="absolute right-0 w-32 flex justify-around" style={{ display: 'flex', alignItems: 'center' }}>
                <div>
                    <span className="text-xs text-neutral-400">{time}</span>
                </div>
                <div>
                    {message.has_thread && <MdOutlineQuestionAnswer size={24} />}
                </div>
            </div>
        </div>
    )
}

export default Message;
