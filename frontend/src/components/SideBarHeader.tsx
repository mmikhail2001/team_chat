import React, { useRef, useContext, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ChannelsContext, ChannelContext } from '../contexts/channelctx';
import Routes from '../config';
import { PopUpContext } from '../contexts/popup';
import CreateChannel from './popup/CreateChannel';

export default function SideBarHeader() {
    const navigate = useNavigate();
    const popup_ctx = useContext(PopUpContext);
    const InviteCode = useRef<HTMLInputElement>(undefined!);
    const channel_context: ChannelContext = useContext(ChannelsContext);

    function JoinChannel(event: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
        event.preventDefault()
        const inv_code = InviteCode.current.value;
        if (inv_code == "") {
            return
        }
        const url = Routes.Invites+`/${inv_code}`
        fetch(url, {
            method: "GET",
            }).then(response => {
                if (response.status === 200) {
                    response.json().then(channel => {
                        channel_context.setChannel(prevChannels => new Map(prevChannels.set(channel.id, channel)));
                        navigate(`/channels/${channel.id}`)
                    })
                }
            })
            // а если липовый invite link, вообще это пользователь никак не поймет
        
        InviteCode.current.value = ""
    }

    return (
        <div className='my-3 relative w-full flex flex-col p-2 border-b border-zinc-800'>
            <input className='h-6 px-2 border-none rounded bg-zinc-300 focus:outline-none' type="text" placeholder="Invite Code" ref={InviteCode} />
            <button className='bg-sky-300 rounded-md h-6 hover:bg-sky-500 my-2' onClick={JoinChannel}>Join Channel</button>
            <button className="bg-sky-300 rounded-md h-6 hover:bg-sky-500" onClick={() => popup_ctx.open(<CreateChannel />) }>Create Channel</button>
        </div>
  )
}
