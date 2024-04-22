import { ChannelContext, ChannelsContext } from '../contexts/channelctx';
import { UserContextOBJ, UserContext } from '../contexts/usercontext';
import { useContext, useEffect, useState, useRef } from 'react';
import { ChannelOBJ, UserOBJ } from '../models/models';
import ChannelList from './channel_list';
import Routes from '../config';
import { useLocation } from 'react-router-dom';
import SideBarHeader from './SideBarHeader';
import { IoClose } from 'react-icons/io5';

function SideBar() {
	const [channels_element, setChannels_element] = useState<JSX.Element[]>([])
	const channel_context: ChannelContext = useContext(ChannelsContext);
	const user_ctx:UserContextOBJ = useContext(UserContext);

	const location = useLocation()

	const searchChats = useRef<HTMLInputElement>(undefined!);
	const [searchValue, setSearchValue] = useState('');
	const [searchResults, setSearchResults] = useState<JSX.Element[]>([]);
	const [isSearchActive, setIsSearchActive] = useState(false);


	useEffect(() => {
		setChannels_element([])
		function sortChannel(a: ChannelOBJ, b: ChannelOBJ) {
			const a_msg = channel_context.messages.get(a.id)
			const b_msg = channel_context.messages.get(b.id)
			if (a_msg && b_msg) {
				const a_msgs = Array.from(a_msg.values()).sort((a, b) => { return a.created_at - b.created_at; });
				const b_msgs = Array.from(b_msg.values()).sort((a, b) => { return a.created_at - b.created_at; });
				if (a_msgs.length > 0 && b_msgs.length > 0) {
					return b_msgs[b_msgs.length - 1].created_at - a_msgs[a_msgs.length - 1].created_at;
				} else if (a_msgs.length > 0) {
					return -1;
				} else if (b_msgs.length > 0) {
					return 1;
				} else {
					return 0;
				}
			} else if (a_msg) {
				return -1
			} else if (b_msg) {
				return 1
			} else {
				return 0
			}
		}

		const channels = Array.from(channel_context.channels.values()).sort(sortChannel)

		channels.forEach(channel => {
			if (channel.type !== 3) {
				setChannels_element(prevElement => [...prevElement,
					<ChannelList key={channel.id} channel={channel} />
				]);
				}
			})
		}, [channel_context.channels, channel_context.messages])

	function handleSearchChange(event: React.ChangeEvent<HTMLInputElement>) {
		console.log("user_ctx.is_guest ============", user_ctx.is_guest)
		const { value } = event.target;
		setSearchValue(value);
		if (value.trim() === '') {
			setIsSearchActive(false);
			return;
		}

		const url = Routes.Search + `?query=${value}&chats=true&employees=true`
		fetch(url, {
			method: 'GET',
			headers: {
				'Content-Type': 'application/json'
			},
		})
			.then(response => {
				if (searchChats.current.value.trim() === '') {
					setIsSearchActive(false);
					return;
				}
				if (response.status === 200) {
					response.json().then(data => {
						const searchResults: JSX.Element[] = [];

						if (data.Channels && data.Channels.length > 0) {
							searchResults.push(
								<div key="channelsDivider" className="px-5 text-gray-500">Channels</div>
							);
							data.Channels.forEach((channel: ChannelOBJ) => {
								searchResults.push(<ChannelList key={channel.id} channel={channel} />);
							});
						}

						if (data.Users && data.Users.length > 0) {
							searchResults.push(
								<hr key="separator" className="my-5"/>,
								<div key="usersDivider" className="px-5 text-gray-500">Other Employees</div>
							);
							data.Users.forEach((user: ChannelOBJ) => {
								searchResults.push(<ChannelList key={user.id} channel={user} />);
							});
						}
						setSearchResults(searchResults);
                    	setIsSearchActive(true);
					});
				} else if (response.status === 404) {
					setSearchResults([<div className='mx-5'>Not found</div>]);
					setIsSearchActive(true);
				}
				else {
					console.log("Search failed or not found");
					setIsSearchActive(false);
				}
			})
	}

	function cancelSearch() {
		setSearchValue('');
		setIsSearchActive(false);
	}

	return (
		<div className={`h-full w-full lg:w-64 ${location.pathname !== "/channels" ? "hidden" : "block"} 
			overflow-y-scroll lg:block md:border-r border-zinc-800`}>
			<SideBarHeader />
			<div className="flex items-center">
				{!user_ctx.is_guest && 
					<input className='h-6 w-44 my-3 mx-5 px-2 border-none rounded bg-zinc-800 focus:outline-none' 
					type="text" placeholder="Search" ref={searchChats} onChange={handleSearchChange} value={searchValue} />
				}
				{searchValue && <IoClose
					className="text-white cursor-pointer"
					onClick={cancelSearch}
					size="20"
					style={{ marginLeft: '-7px', marginTop: '2px' }}
					/>}
			</div>
			
			{!isSearchActive ? (
				<div className="px-5 py-3 my-4 bg-gray-900 text-gray-500">My Chats</div>
			) : (
				<div className="px-5 py-3 my-4 bg-gray-900 text-gray-500">Search Results</div>
			)}
			{isSearchActive ? searchResults : channels_element}
		</div>
	);
}
export default SideBar;
