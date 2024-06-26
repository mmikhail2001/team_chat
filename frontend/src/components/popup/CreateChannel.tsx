import { useContext, useRef, useState } from "react";
import { UserContextOBJ, UserContext } from "../../contexts/usercontext";
import { CreateChannel as APICreateChannel } from "../../api/channel";
import { ChannelContext, ChannelsContext } from "../../contexts/channelctx";
import { HiCamera, HiX } from "react-icons/hi";
import { PopUpContext } from "../../contexts/popup";
import { setDefaultIcon } from "../../utils/errorhandle";

export default function CreateChannel() {
    const popup_ctx = useContext(PopUpContext);
    const user: UserContextOBJ = useContext(UserContext);
    const channel_context: ChannelContext = useContext(ChannelsContext);

    const [isNewsChannel, setIsNewsChannel] = useState(false);
    const channel_name = useRef<HTMLInputElement>(null);
    const icon_input = useRef<HTMLInputElement>(null);
    const icon_image = useRef<HTMLImageElement>(null);

    function HandleCreateChannel(e: React.MouseEvent<Element, MouseEvent>) {
        e.preventDefault();
        const channelName = channel_name.current?.value;
        if (!channelName) return;

        const isNews = isNewsChannel;

        if (icon_input.current?.files && icon_input.current.files.length > 0) {
            let reader = new FileReader();
            reader.readAsDataURL(icon_input.current.files[0]);
            reader.onload = () => {
                APICreateChannel(channelName, reader.result, isNews).then(new_channel => {
                    channel_context.setChannel(p => new Map(p.set(new_channel.id, new_channel)));
                });
            };
        } else {
            APICreateChannel(channelName, "", isNews).then(new_channel => {
                channel_context.setChannel(p => new Map(p.set(new_channel.id, new_channel)));
            });
        }

        popup_ctx.close();
    }

    function onIconChange() {
        if (icon_input.current?.files && icon_input.current.files.length > 0) {
            const file = icon_input.current.files[0];
            if (file.size > 2097152) {
                alert("Image is bigger than 2MB");
                icon_input.current.value = '';
                return;
            }
            if (icon_image.current) {
                icon_image.current.src = URL.createObjectURL(file);
            }
        }
    }

    return (
        <div onClick={(e) => e.stopPropagation()} className='relative rounded-2xl text-black bg-zinc-300 h-96 w-96 flex items-center'>
            <HiX size={24}  onClick={() => popup_ctx.close()} className="absolute top-0 right-0 m-4 cursor-pointer text-slate-400" />
            <div className="flex flex-col items-center w-full">
                <div className="relative flex items-center justify-center h-32 w-32">
                    <img onClick={() => icon_input.current?.click()} onError={setDefaultIcon} className="h-24 w-24 rounded-xl cursor-pointer p-0 m-2 border-slate-700 border-2 border-dashed" ref={icon_image} alt="icon" src="" />
                    <HiCamera size={64} onClick={() => icon_input.current?.click()} className="absolute self-center justify-self-center text-white opacity-75 cursor-pointer" />
                    <input type="file" ref={icon_input} name="filename" hidden onChange={onIconChange} accept="image/*" />
                </div>
                <input className="bg-slate-100 rounded p-2 m-2" ref={channel_name} type="text" placeholder="Channel Name" defaultValue={`${user.username}'s channel`} />
                <div className="flex items-center mb-4">
                    <label htmlFor="newsChannelCheckbox">Create a news channel?</label>
                    <input 
                        type="checkbox" 
                        checked={isNewsChannel} 
                        onChange={(e) => setIsNewsChannel(e.target.checked)} 
                        id="newsChannelCheckbox"
                        className="mr-2 ml-3 my-3"
                    />
                </div>
                <button className="h-10 bg-green-400 w-24 rounded hover:bg-green-500" onClick={HandleCreateChannel}>Create</button>
            </div>
        </div>
    );
}
