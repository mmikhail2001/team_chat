import { useContext, useEffect, useState } from "react";
import { MdOutlineChatBubbleOutline, MdOutlinePeopleAlt } from "react-icons/md";
import { IoMailOpenOutline } from "react-icons/io5";
import { HiOutlineCog } from "react-icons/hi";
import { useLocation, useNavigate } from "react-router-dom";
import { UserContextOBJ, UserContext } from "../contexts/usercontext";
import { setDefaultAvatar } from '../utils/errorhandle';

export default function NavBar() {
    const [hideNavBar, setHideNavBar] = useState(false)
    const location = useLocation()
    useEffect(() => {
        if (/\/channels\/.+/.test(location.pathname)) {
            setHideNavBar(true)
        } else {
            setHideNavBar(false)
        }
    }, [location])
    
    const navigate = useNavigate()
    const user:UserContextOBJ = useContext(UserContext);

    const isActiveIcon = (path: string) => location.pathname.includes(path);
    
    return (
        <div className={`h-12 w-full md:h-full md:w-12 ${ hideNavBar ? 'hidden' : 'flex' } md:flex md:flex-col items-center justify-evenly md:justify-start border-t md:border-t-0 md:border-r border-zinc-800`}>
            <img src={user.avatar} alt="avatar" className="w-8 h-8 m-2 bg-zinc-900 hover:cursor-pointer rounded-xl" onError={setDefaultAvatar} onClick={ () => navigate("/") } />
            <MdOutlineChatBubbleOutline size={32} className={`m-2 hover:opacity-90 ${isActiveIcon('/channels') ? 'opacity-80' : 'opacity-30'}`} onClick={ () => navigate("/channels") } />
            <IoMailOpenOutline size={32} className={`m-2 hover:opacity-90 ${isActiveIcon('/mailings') ? 'opacity-80' : 'opacity-30'}`} onClick={ () => navigate("/mailings") } />
            {/* <MdOutlinePeopleAlt size={32} className={`m-2 hover:opacity-90 ${isActiveIcon('/relationships') ? 'opacity-80' : 'opacity-30'}`} onClick={ () => navigate("/relationships") } /> */}
            <HiOutlineCog size={32} className={`m-2 hover:opacity-90 ${isActiveIcon('/settings') ? 'opacity-80' : 'opacity-30'}`} onClick={ () => navigate("/settings") } />
        </div>
    )
}
