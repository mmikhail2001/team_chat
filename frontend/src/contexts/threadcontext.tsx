import { useState, createContext, useEffect  } from "react";
import { ChannelOBJ, MessageOBJ } from '../models/models';
import { useParams } from 'react-router-dom';


export interface ThreadContextOBJ {
    channel: ChannelOBJ
    setChannel: React.Dispatch<React.SetStateAction<ChannelOBJ>>
    thread: ChannelOBJ
    setThread: React.Dispatch<React.SetStateAction<ChannelOBJ>>
    message: MessageOBJ
    setMessage: React.Dispatch<React.SetStateAction<MessageOBJ>>
    threadShow: boolean
    setThreadShow: React.Dispatch<React.SetStateAction<boolean>>
}

export const ThreadContext = createContext<ThreadContextOBJ>(undefined!);

function ThreadCTX({ children }: { children: React.ReactChild }) {
    const [channel, setChannel] = useState<ChannelOBJ>(undefined!);
    const [thread, setThread] = useState<ChannelOBJ>(undefined!);
    const [message, setMessage] = useState<MessageOBJ>(undefined!);
    const [threadShow, setThreadShow] = useState(false);
    const { id } = useParams<{ id: string }>();

    useEffect(() => {
        if (channel && id !== channel.id) {
            setThreadShow(false);
        }
    }, [id, channel]);

    const value: ThreadContextOBJ = {
        channel,
        setChannel,
        thread,
        setThread,
        message,
        setMessage,
        threadShow,
        setThreadShow
    }

    return (
        <ThreadContext.Provider value={value}>
            {children}
        </ThreadContext.Provider>
    );
}

export default ThreadCTX;
