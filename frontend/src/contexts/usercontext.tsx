import { createContext, useEffect, useState } from "react";
import useMap from '../hooks/useMap';
import Routes from "../config";
import { UserOBJ } from "../models/models";
import { Relationship } from "../models/relationship";
import { GetRelationships } from "../api/relationship";
import { GetMessages } from "../api/message";
import { useNavigate } from "react-router-dom";

export type UserContextOBJ = {
    id: string;
    username: string;
    avatar: string;
    setId:React.Dispatch<React.SetStateAction<string>>;
    setUsername:React.Dispatch<React.SetStateAction<string>>;
    setAvatar:React.Dispatch<React.SetStateAction<string>>;
    relationships: Map<String,Relationship>;
	setRelationships: React.Dispatch<React.SetStateAction<Map<String, Relationship>>>;
    deleterelationship: (key: String) => void;
    
    reactions: Map<String,String>;
	setReactions: React.Dispatch<React.SetStateAction<Map<String, String>>>;
}

export const UserContext = createContext<UserContextOBJ>(undefined!);

function UserCTX({ children }: { children: React.ReactChild }) {
    const navigate = useNavigate();
    const [id, setId] = useState<string>("");
    const [username, setUsername] = useState<string>("");
    const [avatar, setAvatar] = useState<string>("");
	const [relationships, setRelationships, deleterelationship] = useMap<Relationship>(new Map<String,Relationship>());
	const [reactions, setReactions] = useState(new Map<String,String>());

    useEffect(() => {
        fetch(Routes.currentUser).then(response => {
            if (response.status === 200) {
                response.json().then((user: UserOBJ) => {
                    setId(user.id);
                    setUsername(user.username);
                    setAvatar(user.avatar);
                    const reactionsMap = new Map<string, string>();
                    user.reactions.forEach(reaction => {
                        reactionsMap.set(reaction.message_id, reaction.reaction);
                    });
                    setReactions(reactionsMap);
                });

                // запрос всех: друзей, кто онлайн, кто офлайн
                GetRelationships().then(relationships => {
                    relationships.forEach(relationship => {
                        setRelationships(prevRelationships => {
                            prevRelationships.set(relationship.id, relationship);
                            return prevRelationships;
                        });
                    });
                });
            } else {
                if (!location.pathname.includes("auth")) {
                    navigate('/auth/login')
                }
            }
        })
    }, []);

    const context_value: UserContextOBJ = {
        id: id,
        username: username,
        avatar: avatar,
        setId: setId,
        setUsername: setUsername,
        setAvatar: setAvatar,
        relationships: relationships,
        setRelationships: setRelationships,
        deleterelationship: deleterelationship,

        reactions: reactions,
        setReactions: setReactions,
    }
    
    return (
    <UserContext.Provider value={context_value}>
        {children}
    </UserContext.Provider>
    )
}

export default UserCTX;
