import { isImage } from "../../utils/attachments";
import { r } from "../../routes";
import MediaView from "../ui/MediaView";
import FilePreview from "./FilePreview";

export function ImagePreview({ url, alt = "" }) {
  return <MediaView url={url} alt={alt} />;
}

export default function AttachmentView({ attachments }) {
  if (!attachments || attachments.length === 0) return null;
  return (
    <div className="mt-1 flex flex-col gap-1">
      {attachments.map((att) => {
        const url = r.attachment(att.id);
        if (isImage(att)) {
          return <MediaView key={att.id} url={url} alt={att.filename} />;
        }
        return <FilePreview key={att.id} att={att} url={url} />;
      })}
    </div>
  );
}