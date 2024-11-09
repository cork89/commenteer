var commentState = 1

async function takeScreenshot(commenteerUrl) {
    // let data;
    // html2canvas(document.querySelector(".image-comment-container"), { allowTaint: true, useCORS: true }).then(canvas => {
    //     // document.body.appendChild(canvas)
    //     // let link = document.createElement("a");
    //     // document.body.appendChild(link);
    //     // link.download = window.location.pathname.split("/")[2] + ".jpg";
    //     // const jpeg = new Image();
    //     // jpeg.src = canvas.toDataURL("image/webp");
    //     // link.href = jpeg.src;
    //     // link.target = '_blank';
    //     // link.click();
    //     data = canvas.toDataURL("image/webp");
    // });
    const commentContainer = commentState == 1 ? document.querySelector(".image-comment-container") : document.querySelector(".image-comment-container-rel");

    const canvas = await html2canvas(commentContainer, { allowTaint: true, useCORS: true });
    let data = canvas.toDataURL("image/webp");
    const pathname = window.location.pathname.split("/")[2];
    const url = `${commenteerUrl}/r/${pathname}/submit/`;
    const [type, imgData] = data.split(",")
    const headers = new Headers();
    headers.append("Content-Type", "image/webp");
    headers.append("Content-Transfer-Encoding", "base64")
    const res = await fetch(url, {
        headers: headers,
        method: "POST",
        body: imgData,
        redirect: "follow",
        credentials: 'include',
    });
    if (!res.ok) {
        console.log("failed to post image")
    }
    else {
        console.log("success posting")
        window.location.href = `/r/${pathname}`
    }
}

function download(data) {
    let link = document.createElement("a");
    document.body.appendChild(link);
    link.download = url + ".webp";
    link.href = data;
    link.target = '_blank';
    link.click();
}

function overlayComments(newState) {
    if (newState == commentState) {
        return
    }
    const icc = document.getElementById("icc")
    const cc = document.getElementById("cc")
    const cs = document.getElementById(`oc${commentState}`)
    const ns = document.getElementById(`oc${newState}`)

    icc.classList.remove("image-comment-container")
    icc.classList.remove("image-comment-container-rel")
    cc.classList.remove("comment-container")
    cc.classList.remove("comment-container-bot")
    cc.classList.remove("comment-container-top")
    cs.classList.remove("target")

    if (newState == 1) {
        icc.classList.add("image-comment-container")
        cc.classList.add("comment-container")
        ns.classList.add("target")
    }

    if (newState == 2) {
        icc.classList.add("image-comment-container-rel")
        cc.classList.add("comment-container-bot")
        ns.classList.add("target")
    }

    if (newState == 3) {
        icc.classList.add("image-comment-container-rel")
        cc.classList.add("comment-container-top")
        ns.classList.add("target")
    }

    commentState = newState
}