import { useContext } from 'react';
import { MessageOBJ } from '../models/models';
import { ChannelsContext } from "../contexts/channelctx";
import Routes from '../config'
import { HiXMark } from 'react-icons/hi2';

export default function PinnedMessage({ message }: {message: MessageOBJ}) {
    const channel_ctx = useContext(ChannelsContext);

    let date = new Date(message.created_at * 1000).toLocaleDateString();
    let time = new Date(message.created_at * 1000).toLocaleTimeString();


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

    return (
        <div className='w-11/12 bg-slate-200 mb-4 rounded p-2'>
            <div className='flex items-center justify-between'>
                <div className='flex items-center rounded-md p-2 bg-slate-300'><h4>{message.author.username}</h4> <p className='text-xs mx-4 text-zinc-800'>{time} - {date}</p></div>
                <button className='bg-neutral-200 rounded-full' onClick={UnpinMsg}>
                    <HiXMark />
                </button>
            </div>
            <div className='break-words'>
                <span>{message.content}</span>
            </div>
        </div>
    )
}
