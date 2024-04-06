import React, { useState, useContext, useRef, useMemo, useEffect } from 'react';
import Message from './message';
import { MessageOBJ, ChannelOBJ } from '../../models/models';
import { ChannelsContext, ChannelContext } from "../../contexts/channelctx";
import { ThreadContext, ThreadContextOBJ } from "../../contexts/threadcontext";
import Routes from '../../config';
import { IoClose } from 'react-icons/io5';

function Thread() {
	// Emoji picker https://www.cluemediator.com/how-to-add-emoji-picker-in-the-react
	const channel_context: ChannelContext = useContext(ChannelsContext);
	const thread_context: ThreadContextOBJ = useContext(ThreadContext);
	// thread_id, если есть
	const channel_id = thread_context.thread.id

	const [Input_message, setInput_message] = useState('');
	const messagesContainerRef = useRef<HTMLDivElement>(null);

	const messageElement = useRef<HTMLDivElement>(null);

	// const channel: ChannelOBJ = channel_context.channels.get(channel_id) || {} as ChannelOBJ;

	// TODO: по веб-сокету придет обновление о новых сообщениях, тогда перерендерим MessageElement
	const MessageElement = useMemo(() => {
		let messagesList: JSX.Element[] = [];

		let messages = channel_context.messages.get(channel_id) || []
		// let messages = channel_context.messages.get(channel_id) || [];

		if (!messages) {
			messages = [] as MessageOBJ[]
		}

		let preDate: string = ""
		let prevAuthor: string = ""
		messages.forEach((message, index) => {
			let date = new Date(message.created_at * 1000).toLocaleDateString();
			let short = prevAuthor === message.author.id;
			prevAuthor = message.author.id;

			const isLastMessage = index === messages.length - 1;

			if (preDate !== date) {
				messagesList.push(
					<div key={date} className="relative flex items-center justify-center h-8">
						<span className='absolute w-full border-t-2 border-zinc-800'></span>
						<span className='absolute bg-black px-4'>{date}</span>
					</div>
				);
				preDate = date;
				short = false;
			}
			messagesList.push(
				<div ref={isLastMessage ? messageElement : null}>
					<Message key={message.id} message={message} short={short} />
				</div>
			)
		});

		return messagesList;
	}, [channel_context.messages, channel_id]);

	useEffect(() => {
		if (messageElement.current !== null) {
			messageElement.current.scrollIntoView({
				behavior: "smooth",
				block: "end"
			});
		}
	}, [channel_context.messages, channel_id]);

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
			if (Input_message.length > 0) {
				const url = Routes.Channels + "/" + channel_id + "/messages";
				fetch(url, {
					method: "POST",
					headers: {
						"Content-Type": "application/json"
					},
					body: JSON.stringify({ content: Input_message })
				})
			}
			setInput_message('');
		}
	}

	function unshowThread() {
		thread_context.setThreadShow(false)
	}
	return (
		<div className="relative h-full w-1/2 flex-col flex">
			<div className='absolute w-full h-16 flex items-center px-6 bg-slate-800'>
				<div className='w-full flex items-center justify-between'>
					<span className='text-lg p-2'>
						Thread by message:  <span className='text-lg p-2 rounded-md bg-slate-900'>{thread_context.message.content}</span>
					</span> 
					<div className="flex items-center">
						<IoClose
							className="text-white cursor-pointer"
							onClick={unshowThread}
							size="20"
							style={{ marginLeft: '-7px', marginTop: '2px' }}
						/>
					</div>
				</div>
			</div>
	
			<div className='flex mt-16 h-full overflow-hidden w-full'>
				<div className='flex-col flex relative w-full'>
					<div className="mb-16 flex-col-reverse overflow-x-hidden overflow-y-scroll">
						{MessageElement}
					</div>
					<div className="h-16 absolute bottom-0 w-full flex items-center justify-evenly border-t border-zinc-800">
						<input className='w-[85%] h-8 rounded-md bg-zinc-800 px-4' type="text" placeholder="Type a message..." onKeyPress={updateChat} value={Input_message} onChange={onInputChange} />
					</div>
				</div>
			</div>
		</div>
	);
	
}

export default Thread;
