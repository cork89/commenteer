async function likePost(event, commenteerUrl, queryId) {
    const url = `${commenteerUrl}/r/${queryId}/like/`
    const res = await fetch(url, {
        method: "POST",
    })

    if (!res.ok) {
        console.log("failed to like")
    } else {
        event.setAttribute("fill", event.getAttribute("fill") == "#FB7185" ? "none" : "#FB7185")
        event.setAttribute("stroke", event.getAttribute("stroke") == "#FB7185" ? "#57534E" : "#FB7185")
    }
}

function login() {
    window.location.href = "/login/"
}

function base64ToBlob(base64, mimeType = "image/png") {
    const byteCharacters = atob(base64.split(',')[1]);

    const byteArray = new Uint8Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
        byteArray[i] = byteCharacters.charCodeAt(i);
    }
    return new Blob([byteArray], { type: mimeType });
}

async function copyImage(event, commenteerUrl) {
    try {
        const queryId = event.id.substr(0, event.id.indexOf("-copy"))
        const imgDiv = document.getElementById(queryId)
        const canvas = await html2canvas(imgDiv, { allowTaint: true, useCORS: true, height: imgDiv.height, width: imgDiv.width, scale: 2 });
        const imgData = canvas.toDataURL("image/png");
        const imgBlob = base64ToBlob(imgData)

        navigator.clipboard.write([
            new ClipboardItem({
                'image/png': imgBlob
            })
        ]);
        const clipboardNote = document.getElementById("clipboard-note")
        clipboardNote.classList.remove("invisible")
        setTimeout(() => {
            clipboardNote.classList.add("invisible")
        }, 5000)
    } catch (error) {
        console.error(error);
    }
}


async function downloadImage(event, commenteerUrl) {
    try {
        const queryId = event.id.substr(0, event.id.indexOf("-download"))
        const imgDiv = document.getElementById(queryId)
        const canvas = await html2canvas(imgDiv, { allowTaint: true, useCORS: true, height: imgDiv.height, width: imgDiv.width, scale: 2 });
        let link = document.createElement("a");
        document.body.appendChild(link);
        link.href = canvas.toDataURL("image/webp");
        link.download = `${queryId}.webp`;
        link.click();
        document.body.removeChild(link)

    } catch (error) {
        console.error(error);
    }
}
