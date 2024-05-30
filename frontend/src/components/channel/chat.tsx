import Picker, { IEmojiData } from 'emoji-picker-react';
import React, { useState, useContext, useLayoutEffect, useRef, useMemo, useEffect } from 'react';
import Message from './message';
import Header from './header';
import { MessageOBJ, ChannelOBJ } from '../../models/models';
import { ChannelsContext, ChannelContext } from "../../contexts/channelctx";
import { ThreadContext, ThreadContextOBJ } from "../../contexts/threadcontext";
import { UserContextOBJ, UserContext } from "../../contexts/usercontext";
import { BsPlusCircleFill } from 'react-icons/bs';
import Routes from '../../config';
import { useParams } from 'react-router-dom';
import Recipients from './Recipients';
import Thread from './thread';
import { FaFile } from 'react-icons/fa';
import { HiXMark } from 'react-icons/hi2';

function Chat() {
	const parameter = useParams<string>();
	let channel_id = parameter.id || "";

	// Emoji picker https://www.cluemediator.com/how-to-add-emoji-picker-in-the-react
	const channel_context: ChannelContext = useContext(ChannelsContext);
	const user_ctx: UserContextOBJ = useContext(UserContext);
	const thread_context: ThreadContextOBJ = useContext(ThreadContext);

	const [Input_message, setInput_message] = useState('');
	const [showPicker, setShowPicker] = useState(false);
	const [showRecipients, setShowRecipients] = useState(true);
	const messagesContainerRef = useRef<HTMLDivElement>(null);

	const messageElement = useRef<HTMLDivElement>(null);

	// const channel: ChannelOBJ = channel_context.channels.get(channel_id) || {} as ChannelOBJ;
	const [channel, setChannel] = useState<ChannelOBJ>({} as ChannelOBJ);

	useEffect(() => {
        const currentChannel = channel_context.channels.get(channel_id) || {} as ChannelOBJ;
        if (Object.keys(currentChannel).length === 0) {
            // Если channel пустой, делаем запрос
            fetch(`/api/channels/${channel_id}`)
                .then(response => response.json())
                .then(data => {
                    setChannel(data);
                    // Обновите контекст каналов, если это необходимо
                    // channel_context.setChannels(prev => new Map(prev).set(channel_id, data));
                })
                .catch(error => {
                    console.error('Error fetching channel:', error);
                });
        } else {
            setChannel(currentChannel);
        }
    }, [channel_id]);

	// TODO: по веб-сокету придет обновление о новых сообщениях, тогда перерендерим MessageElement
	const MessageElement = useMemo(() => {
		let messagesList: JSX.Element[] = [];

		let messages = channel_context.messages.get(channel_id) || [];

		if (!messages) {
			messages = [] as MessageOBJ[]
		}

		let preDate: string = ""
		let prevAuthor: string = ""
		messages.forEach((message, index) => {
			let date = new Date(message.created_at * 1000).toLocaleDateString();
			let short;
			if (channel.type == 4) {
				short = false;
			} else {
				short = prevAuthor === message.author.id;
			}
			prevAuthor = message.author.id;

			const isLastMessage = index === messages.length - 1;
			
			if (preDate !== date) {
				messagesList.push(
					<div key={date} className="relative flex items-center justify-center h-8">
						<span className='absolute w-full border-t-2 border-zinc-300'></span>
						<span className='absolute bg-zinc-300 rounded-md px-4'>{date}</span>
					</div>
				);
				preDate = date;
				short = false;
			}
			// if, чтобы убрать системные сообщения на новостных каналах и в чат ботах
			if (!(message.system_message && (channel.type == 4 || channel.type == 5))) {
				messagesList.push(
					<div ref={isLastMessage ? messageElement : null}>
						<Message key={message.id} message={message} short={short} />
					</div>
					)
			}
			// messagesList.push(<Message key={message.id} isRef={isLastMessage} message={message} short={short} />)
		});

		return messagesList;
	}, [channel_context.messages, channel_id]);

	const [hasFile, setHasFile] = useState(false);
	const file_input = useRef<HTMLInputElement>(undefined!);
	const [fileJSX, setFileJSX] = useState<JSX.Element>(<></>);

	// TODO: надо добавить
	const onEmojiClick = (_: React.MouseEvent<Element, MouseEvent>, data: IEmojiData) => {
		setInput_message(prevInput => prevInput + data.emoji);
		setShowPicker(false);
	};

	useEffect(() => {
		console.log('useEffect def')
    }, []);

	// TODO: не может доскролить до конца...
	useEffect(() => {
        if (messageElement.current !== null) {
            messageElement.current.scrollIntoView({
                behavior: "smooth",
                block: "end"
            });
        }
    }, [channel_context.messages, channel_id]);

	// useLayoutEffect(() => {
	// 	console.log('useEffect channel_context')
	// 	if (messagesContainerRef.current !== null) {
	// 	  const { scrollHeight, clientHeight } = messagesContainerRef.current;
	// 	  messagesContainerRef.current.scrollTop = scrollHeight - clientHeight;
	// 	}
	//   }, [channel_context.messages, channel_id]);

	function onInputChange(event: React.ChangeEvent<HTMLInputElement>) {
		const inputstr = event.target.value;
		if (inputstr.length <= 2000) {
			setInput_message(inputstr);
		} else {
			alert("Message too long");
		}
	}
	function updateChat(event: React.KeyboardEvent<HTMLInputElement>) {
		if (event.key === 'Enter') {
			event.preventDefault();
			if (Input_message.length > 0 && (file_input === null || file_input.current.files?.length === 0)) {
				const url = Routes.Channels + "/" + channel_id + "/messages";
				fetch(url, {
					method: "POST",
					headers: {
						"Content-Type": "application/json"
					},
					body: JSON.stringify({ content: Input_message })
				})
			}

			if (file_input.current.files && file_input.current.files.length > 0) {
				const url = Routes.Channels + "/" + channel_id + "/messages";
				const formData = new FormData();
				formData.append('content', Input_message);
				formData.append('file', file_input.current.files[0]);
				fetch(url, {
					method: "POST",
					body: formData
				})
			}
			setInput_message('');
			setHasFile(false);
			file_input.current.value = '';
		}
	}

	// TODO: почему бы не сделать это отдельным компонентом?
	const onFileChange = () => {
		if (file_input.current.files && file_input.current.files.length > 0) {
			const file = file_input.current.files[0];
			if (file.size > 8388608) {
				alert("File is bigger than 8MB")
				file_input.current.value = ''
				return
			}
			setHasFile(true);
			
			setFileJSX(
				<div className='relative h-32 w-32 mx-4 bg-zinc-300 rounded flex items-center justify-center' key={file.name}>
					<FaFile size={48} />
					<button className='absolute top-1 right-1 bg-none border-none text-red-600' onClick={() => { file_input.current.value = ''; onFileChange(); }}>
						<HiXMark size={20} />
					</button>
					<p className='absolute bottom-1 left-1 m-0 w-28 whitespace-nowrap overflow-hidden text-ellipsis text-xs'>{file.name}</p>
				</div>
			)
		} else {
			setHasFile(false);
		}
	}

	return (
		<div className="relative h-full w-full flex-col flex">
			<Header channel={channel} toggleRecipients={setShowRecipients} />
			<div className='flex mt-16 h-full overflow-hidden w-full'>
				<div className='flex-col flex relative w-full'>
					<div className="mb-16 flex-col-reverse overflow-x-hidden overflow-y-scroll" ref={messagesContainerRef}>
						{MessageElement}
					</div>
					{
						(channel.type == 1 || channel.type == 2 || (channel.type == 5 && channel.owner_id == user_ctx.id )) && 
						<div className="h-16 absolute bottom-0 w-full flex items-center justify-evenly border-t border-zinc-300">
							{ hasFile && <div className='absolute bottom-16 right-0 h-40 w-full flex items-center rounded-t-xl border-t border-r border-l border-zinc-300'>{fileJSX}</div> } 
							<input type="file" ref={file_input} name="filename" hidden onChange={onFileChange} />
							<BsPlusCircleFill color="gray" size={26} onClick={() => file_input.current.click()} />
							<input className='w-[85%] h-8 rounded-md bg-zinc-300 px-4' type="text" placeholder="Type a message..." onKeyPress={updateChat} value={Input_message} onChange={onInputChange} />
						</div>
					}

				</div>
				{ thread_context.threadShow && <Thread/> }
				{ (channel.type === 2 || channel.type === 4 || channel.type === 5) && showRecipients && <Recipients channel={channel} /> }
			</div>
		</div>
	);
}

export default Chat;
