import { Link } from "react-router-dom";

export default function HomeDefault() {
    return <div className="flex w-full h-full items-center justify-center">
        <div className="flex flex-col w-96 items-center">
            <span className="font-semibold">Home</span>
            <span>TeamChat is an open-source application crafted with a ReactJS frontend and a Golang backend, 
                showcasing a robust combination of cutting-edge technologies.
                <br />
                <br />Source code: <Link className="text-cyan-400" to="https://github.com/mmikhail2001/team_chat">github.com/mohanavel15/Chatapp</Link>
            </span>
        </div>
    </div>
}
