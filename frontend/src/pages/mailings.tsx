import { useState, useContext, useEffect } from 'react';
import { UserContextOBJ, UserContext } from "../contexts/usercontext";
import { ChannelsContext, ChannelContext } from "../contexts/channelctx";
import { AddRecipient, RemoveRecipient } from '../api/recipient';
import { ChannelOBJ } from "../models/models";

function Mailings() {
    const user_ctx: UserContextOBJ = useContext(UserContext);
    const channel_context: ChannelContext = useContext(ChannelsContext);
    const [mailings, setMailings] = useState<ChannelOBJ[]>([]);

    useEffect(() => {
        fetch('/api/users/@me/channels/mailings')
            .then(response => response.json())
            .then(data => setMailings(data))
            .catch(error => console.error('Error fetching mailings:', error));
    }, []);

	const handleSubscription = (channel: ChannelOBJ) => {
		const isSubscribed = channel_context.channels.has(channel.id);
	
		if (!isSubscribed) {
			AddRecipient(channel.id, user_ctx.id)
				.then(() => {
					const updatedChannels = new Map(channel_context.channels);
					updatedChannels.set(channel.id, channel);
					channel_context.setChannel(updatedChannels);
				})
				.catch(error => console.error('Error subscribing to the channel:', error));
		} else {
			RemoveRecipient(channel.id, user_ctx.id, "", false)
				.then(() => {
					const updatedChannels = new Map(channel_context.channels);
					updatedChannels.delete(channel.id);
					channel_context.setChannel(updatedChannels);
				})
				.catch(error => console.error('Error unsubscribing from the channel:', error));
		}
	};

    return (
		<div className="w-full flex justify-center">
			<div className="w-96">
				{mailings.map((mailing: ChannelOBJ) => (
					<div key={mailing.id} className="my-10 bg-gray-800 rounded-lg p-4 flex items-center justify-between">
						<div className="flex items-center">
							<img className="w-16 h-16 rounded-full mr-4" src={mailing.icon} />
							<h2 className="text-lg font-semibold">{mailing.name}</h2>
						</div>
						<button
							className={`px-4 py-2 rounded-md font-semibold ${
								channel_context.channels.has(mailing.id) ? 'bg-red-500 text-white' : 'bg-blue-500 text-white'
							}`}
							onClick={() => handleSubscription(mailing)}
						>
							{channel_context.channels.has(mailing.id) ? "Unsubscribe" : "Subscribe"}
						</button>
					</div>
				))}
			</div>
		</div>
	);
}

export default Mailings;
