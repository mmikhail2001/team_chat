import { MessageOBJ } from '../../../models/models';
import { FaDownload } from "react-icons/fa";

export default function AttachmentImage({ message }: { message: MessageOBJ }) {
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
      <div style={{ position: 'relative' }}>
          <img width={"32%"} src={message.attachments[0].url} alt={message.attachments[0].filename} />
          <button className="text-zinc-500 m-3 hover:text-zinc-400 absolute top-0 right-0 m-2" onClick={handleDownload}>
              <FaDownload size={16} />
          </button>
      </div>
  );
}
