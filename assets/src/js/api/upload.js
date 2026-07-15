export function uploadFile(file, onProgress) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const form = new FormData();
    form.append("file", file);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) onProgress(e.loaded / e.total);
    };

    xhr.onload = () => {
      if (xhr.status !== 200) { reject(new Error("Upload failed")); return; }
      try { resolve(JSON.parse(xhr.responseText)); }
      catch { reject(new Error("Bad response")); }
    };

    xhr.onerror = () => reject(new Error("Network error"));

    xhr.open("POST", "/api/v1/attachments");
    xhr.setRequestHeader("Authorization", `Bearer ${window.__token || ""}`);
    xhr.send(form);
  });
}