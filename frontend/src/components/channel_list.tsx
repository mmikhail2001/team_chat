import { Link, useParams } from "react-router-dom";
import { setDefaultIcon, setDefaultAvatar } from '../utils/errorhandle';
import { ChannelOBJ } from "../models/models";
import { UserContextOBJ, UserContext } from "../contexts/usercontext";
import { RxDot, RxDotFilled } from "react-icons/rx";
import { useContext } from "react";
import { ContextMenu } from "../contexts/context_menu_ctx";
import ChannelContextMenu from "../contextmenu/channel_context_menu";

// TODO: может быть это ChannelElement ???
// ведь это не список, а один канал...
// или под лист имеется в виду "лист дерева" ......
export default function ChannelList({ channel }: { channel: ChannelOBJ }) {
    const user_ctx: UserContextOBJ = useContext(UserContext);
    const parameter = useParams<string>();

    let isActive = parameter.id === channel.id;
    const isChannel = channel.type === 2 || channel.type === 4 || channel.type === 5;

    let icon: string;
    let name: string;
    let alt: string;
    let defaultIcon: (event: React.SyntheticEvent<HTMLImageElement, Event>) => void;

    if (isChannel) {
        icon = channel.icon;
        name = channel.name;
        alt = "Icon";
        defaultIcon = setDefaultIcon;
    } else {
        icon = channel.recipients[0].avatar;
        name = channel.recipients[0].username;
        alt = "Avatar";
        defaultIcon = setDefaultAvatar;
    }
    const ctx_menu = useContext(ContextMenu);

    return (
        <Link to={`/channels/${channel.id}`} className="linktag" onContextMenu={(event) => {
            event.preventDefault();
            ctx_menu.open(<ChannelContextMenu x={event.clientX} y={event.clientY} channel={channel} />);
        }}>
            <div className={`w-full h-12 px-2 mt-2 flex items-center cursor-pointer rounded ${isActive && 'bg-zinc-200'} hover:bg-zinc-100`}>
                <div className='relative h-10 w-10 mx-4'>
                    <img className='rounded-full h-10 w-10 bg-zinc-400' style={{ objectFit: 'cover', objectPosition: 'center' }} src={icon} onError={defaultIcon} alt={alt} />
                    {!isChannel && <div className='absolute right-0 bg-zinc-500 rounded-full bottom-0'>
                        {channel.recipients[0].status === 1 ? <RxDotFilled size={17} className="text-green-600" /> : <RxDot size={20} className="text-gray-400" />}
                    </div>}
                </div>
                <p className="w-28 h-6 overflow-hidden text-ellipsis whitespace-nowrap">{name}</p>
                <p className="bg-gray-300 p-1 rounded text-xs ml-auto">
                    {channel.type === 1 ? "Direct" : channel.type === 2 ? "Group" : channel.type === 4 ? "Bot" : channel.type === 5 ? "News" : ""}
                </p>
            </div>
        </Link>
    )
}
