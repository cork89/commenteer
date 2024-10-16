

async function takeScreenshot() {
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
    const canvas = await html2canvas(document.querySelector(".image-comment-container"), { allowTaint: true, useCORS: true });
    let data = canvas.toDataURL("image/webp");
    const pathname = window.location.pathname.split("/")[2];
    const url = `http://localhost:8090/r/${pathname}/submit/`;
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