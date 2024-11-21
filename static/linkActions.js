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

async function copyImage(event, commenteerUrl) {
    try {
        const queryId = event.id.substr(0, event.id.indexOf("-copy"))
        const imgDiv = document.getElementById(queryId)
        const img = imgDiv.src

        // let imgData = sessionStorage.getItem(imgId);
        // if (!imgBlob) {
        const imgData = await fetch(`${commenteerUrl}/image/?src=${img}`, {
            mode: "no-cors",
            headers: {
                "Access-Control-Allow-Origin": "*",
            }
        }).then(res => res.arrayBuffer());
        // sessionStorage.setItem(imgId, imgData);
        // }
        const imgBlob = new Blob([await imgData], { type: 'image/png' });

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
