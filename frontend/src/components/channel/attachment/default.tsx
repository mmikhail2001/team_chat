import { MessageOBJ } from '../../../models/models';
import { FaDownload, FaFile } from "react-icons/fa";

export default function AttachmentDefault({ message }: { message: MessageOBJ }) {
    const handleDownload = () => {
        const url = message.attachments[0].url;
        const filename = message.attachments[0].filename;

        fetch(url)
            .then(response => response.blob())
            .then(blob => {
                const url = window.URL.createObjectURL(new Blob([blob]));
                const link = document.createElement('a');
                link.href = url;
                link.setAttribute('download', filename);
                document.body.appendChild(link);
                link.click();
                link.parentNode?.removeChild(link);
            })
            .catch(error => console.error('Error downloading file:', error));
    };

    return (
        <div className="relative h-16 flex bg-zinc-900 rounded-lg items-center mb-4 p-4">
            <FaFile size={32} />
            <div className="flex flex-col px-4">
                <p className="m-0 text-cyan-400 hover:underline">
                    <a href={message.attachments[0].url} rel="noreferrer" target="_blank">
                        {message.attachments[0].filename}
                    </a>
                </p>
                <p className="m-0 text-xs">{message.attachments[0].size} bytes</p>
            </div>
            {/* <button className="absolute right-4 text-zinc-500 hover:text-zinc-400" onClick={handleDownload}> */}
            <button className="absolute m-3 top-0 right-0 text-zinc-500 hover:text-zinc-400" onClick={handleDownload}>
                <FaDownload className="" size={16} />
            </button>
        </div>
    );
}
