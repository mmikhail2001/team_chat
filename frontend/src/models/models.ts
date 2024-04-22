export type MessageOBJ = {
    id: string
    content: string
    author: UserOBJ
    channel_id: string
    thread_id: string
    has_thread: boolean
    created_at: number
    updated_at: number
    attachments: Attachment[]
    reactions: ReactionOBJ[]
    system_message: boolean
}

export type ReactionOBJ = {
    user_id: string
    reaction: string
}

export type UserOBJ = {
    id: string
    username: string
    avatar: string
    // онлайн, не онлайн
    status: number
    created_at: number
    reactions: ReactionMessageOBJ[]
    is_guest: boolean
}

export type ReactionMessageOBJ = {
    message_id: string
    reaction: string
}

// type User struct {
// 	ID        string            `json:"id"`
// 	Avatar    string            `json:"avatar"`
// 	Username  string            `json:"username"`
// 	Status    int               `json:"status"`
// 	CreatedAt int64             `json:"created_at"`
// 	Reactions []ReactionMessage `json:"reactions"`
// }

// type ReactionMessage struct {
// 	MessageID string `json:"message_id"`
// 	Reaction  string `json:"reaction"`
// }


export type Attachment = {
    id: string
    filename: string
    size: number
    content_type: string
    url: string
}

export type WS_Message = {
    event: string
    data: any
}

export type Msg_request = {
    content: string
    channel: string
}


export type InviteOBJ = {
    invite_code: string
    created_at: string
}

export type BanOBJ = {
    id: string
    banned_by: UserOBJ
    banned_user: UserOBJ
    channel: ChannelOBJ
    reason: string
    created_at: number
}

export type ChannelOBJ = {
    id: string
    name: string
    icon: string
    type: number
    owner_id: string
    created_at: string
    recipients: UserOBJ[]
}
export type ReadyOBJ = {
    user: UserOBJ
    channels: ChannelOBJ[]
}

export type Status = {
    user_id: string
    status: number
    type: number
    channel_id: string
}
