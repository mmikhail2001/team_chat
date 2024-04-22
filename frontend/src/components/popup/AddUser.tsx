import { useContext, useEffect, useState } from 'react';
import { PopUpContext } from '../../contexts/popup';
import { AddRecipient, AddRoleToChannel } from '../../api/recipient';
import { setDefaultAvatar } from '../../utils/errorhandle';
import { ChannelOBJ } from '../../models/models';
import Routes from '../../config';
import { ChannelContext, ChannelsContext } from '../../contexts/channelctx';
import { FaPeopleGroup } from "react-icons/fa6";

export default function AddUser({ id }: { id: string }) {
    const popup_ctx = useContext(PopUpContext);
    const channel_context: ChannelContext = useContext(ChannelsContext);

    const [elements, setElements] = useState<JSX.Element[]>([]);
    const [username, setUsername] = useState<string>("");

    useEffect(() => {
        setElements([]);
        const url = Routes.Search + `?query=${username}&chats=false&employees=true&roles=true`;
        fetch(url)
            .then(response => {
                if (response.ok) {
                    return response.json();
                }
                throw new Error('Network response was not ok.');
            })
            .then(data => {
                const users: ChannelOBJ[] = data.Users;
                const roles: string[] = data.Roles;
                let userElements: JSX.Element[] = [];
                if (users && users.length !== 0) {
                    userElements = users.reduce((acc: JSX.Element[], user: ChannelOBJ) => {
                        const channelRecipients = channel_context.channels.get(id)?.recipients;
                        const isRecipientExist = channelRecipients ? channelRecipients.some(recipient => recipient.id === user.recipients[0].id) : false;
                        if (!isRecipientExist) {
                            acc.push(
                                <div className='flex w-full mb-2 items-center relative m-1 bg-gray-800 rounded p-2' key={user.recipients[0].id}>
                                    <img src={user.recipients[0].avatar} className="h-8 w-8 rounded" alt="avatar" onError={setDefaultAvatar} />
                                    <p className='mx-4 text-xl'>{user.recipients[0].username}</p>
                                    <button className='absolute m-2 right-0 h-8 rounded hover:bg-green-600 px-2 border-green-600 border-2' onClick={() => AddRecipient(id, user.recipients[0].id).then(popup_ctx.close)}>Add</button>
                                </div>
                            );
                        }
                        return acc;
                    }, []);
                }
                
                if (roles && roles.length !== 0) {
                    roles.forEach(role => {
                        userElements.push(
                            <div className='flex w-full mb-2 items-center relative m-1 bg-gray-800 rounded p-2' key={role}>
                                <FaPeopleGroup className="h-8 w-8 rounded" />
                                <div className="flex flex-col mx-4">
                                    <p className='text-xl'>{role}</p>
                                    <p className='text-sm text-gray-500'>Role</p>
                                </div>
                                <button className='absolute m-2 right-0 h-8 rounded hover:bg-green-600 px-2 border-green-600 border-2' onClick={() => AddRoleToChannel(id, role).then(popup_ctx.close)}>Add</button>
                            </div>
                        );
                    });
                }
                
                setElements(userElements);
            })
            .catch(error => {
                console.error('There was an error!', error);
            });
    }, [username]);

    return (
        // <div className='relative rounded-2xl text-white bg-zinc-900 h-96 w-96 flex flex-col items-center p-6' onClick={e => e.stopPropagation()} defaultValue={username}>
        //     <input type="text" className='w-full bg-zinc-800 p-2 rounded-md' placeholder='username' onChange={e => setUsername(e.currentTarget.value)} />
        //     <div className='bg-zinc-800 w-full flex flex-col p-4 h-full rounded-md mt-6 overflow-y-scroll'>
        //         {elements}
        //     </div>
        // </div>

        <div className='relative rounded-2xl text-white bg-zinc-900 h-1/2 w-1/4 flex flex-col items-center p-6' onClick={e => e.stopPropagation()} defaultValue={username}>
            <input type="text" className='w-full bg-zinc-800 p-2 rounded-md' placeholder='username or role' onChange={e => setUsername(e.currentTarget.value)} />
            <div className='bg-zinc-800 w-full flex flex-col p-4 h-full rounded-md mt-6 overflow-y-scroll'>
                {elements}
            </div>
        </div>
    );
}
