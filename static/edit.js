var commentState = 1
var borderState = 1
var borderColorState = 1
var fontState = 1
var boldState = false
var italicState = false

/**
 * @param {string} commenteerUrl 
 */
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
    const publishButton = document.getElementById("publish")
    publishButton.classList.add("disabled")
    const commentContainer = commentState == 1 ? document.querySelector(".image-comment-container") : document.querySelector(".image-comment-container-rel");

    const canvas = await html2canvas(commentContainer, { allowTaint: true, useCORS: true });
    let data = canvas.toDataURL("image/webp");
    const pathname = window.location.pathname.split("/")[2]
    const url = `${commenteerUrl}/r/${pathname}/submit/`;
    const [type, imgData] = data.split(",")
    const headers = new Headers();
    headers.append("Content-Type", "application/json")
    // headers.append("Content-Transfer-Encoding", "base64")
    const body = {
        imgData: imgData,
        height: canvas.height,
        width: canvas.width,
        params: window.location.search,
    }
    const res = await fetch(url, {
        headers: headers,
        method: "POST",
        body: JSON.stringify(body),
        redirect: "follow",
        credentials: 'include',
    });
    if (!res.ok) {
        console.log("failed to post image")
        publishButton.classList.remove("disabled")
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

/**
 * @param {number} newState
 */
function initState(newState) {
    commentState = newState
    console.log("initState", newState)
}

/**
 * @param {number} newState
 */
function overlayComments(newState) {
    if (newState == commentState) {
        return
    }
    const params = new URLSearchParams(window.location.search)
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
        params.set("cmt", "outer")
    }

    if (newState == 2) {
        icc.classList.add("image-comment-container-rel")
        cc.classList.add("comment-container-bot")
        ns.classList.add("target")
        params.set("cmt", "bottom")
    }

    if (newState == 3) {
        icc.classList.add("image-comment-container-rel")
        cc.classList.add("comment-container-top")
        ns.classList.add("target")
        params.set("cmt", "top")
    }

    commentState = newState
    window.history.pushState(null, "", `?${params.toString()}`)
}

/**
 * @param {number} newState
 */
function borderWidth(newState) {
    if (newState == borderState) {
        return
    }
    const params = new URLSearchParams(window.location.search)
    const icc = document.getElementById("icc")
    const bs1 = document.getElementById(`bw1`)
    const bs2 = document.getElementById(`bw2`)
    const bs3 = document.getElementById(`bw3`)
    const ns = document.getElementById(`bw${newState}`)

    icc.classList.remove("border-2", "border-4", "border-8")
    bs1.classList.remove("target")
    bs2.classList.remove("target")
    bs3.classList.remove("target")

    if (newState == 1) {
        icc.classList.add("border-2")
        ns.classList.add("target")
        params.set("brd", "2")
    }

    if (newState == 2) {
        icc.classList.add("border-4")
        ns.classList.add("target")
        params.set("brd", "4")
    }

    if (newState == 3) {
        icc.classList.add("border-8")
        ns.classList.add("target")
        params.set("brd", "8")
    }

    borderState = newState
    window.history.pushState(null, "", `?${params.toString()}`)
}

const borderColors = { 1: ["border-slate-500", "#64748B"], 2: ["border-red-700", "#B91C1C"], 3: ["border-blue-700", "#1D4ED8"], 4: ["border-emerald-700", "#047857"], 5: ["border-yellow-500", "#EAB308"] }

/**
 * @param {number} newState
 */
function borderColor(newState) {
    if (newState == borderColorState) {
        closeBorderColors()
        return
    }
    const params = new URLSearchParams(window.location.search)
    const icc = document.getElementById("icc")
    const bc = document.getElementById("border-color")

    icc.classList.remove("border-slate-500", "border-red-700", "border-blue-700", "border-emerald-700", "border-yellow-500")
    icc.classList.add(borderColors[newState][0])
    bc.childNodes[1].setAttribute("fill", borderColors[newState][1])
    params.set("bc", newState)
    borderColorState = newState
    closeBorderColors()
    window.history.pushState(null, "", `?${params.toString()}`)

}

function closeBorderColors() {
    const bcContainer = document.getElementById("bc-container")
    if (!bcContainer.classList.contains("invisible")) {
        bcContainer.classList.add("invisible")
    }
}

function openBorderColors() {
    const bcContainer = document.getElementById("bc-container")
    if (bcContainer.classList.contains("invisible")) {
        bcContainer.classList.remove("invisible")
    } else {
        closeBorderColors()
    }
}


const fonts = { 1: "Arial", 2: "'Brush Script MT', cursive", 3: "'Verdana', sans-serif", 4: "'Times New Roman', serif" }

/**
 * @param {number} newState
 */
function changeFont(event) {
    const newState = event.selectedIndex + 1
    if (newState == fontState) {
        return
    }
    const params = new URLSearchParams(window.location.search)
    const icc = document.getElementById("icc")
    icc.style.fontFamily = fonts[newState]
    params.set("font", newState)
    fontState = event.selectedIndex + 1
    window.history.pushState(null, "", `?${params.toString()}`)
}

function boldFont() {
    newState = !boldState
    const params = new URLSearchParams(window.location.search)
    const icc = document.getElementById("icc")
    const bold = document.getElementById("font-bold")
    if (newState == true) {
        bold.classList.add("target")
        icc.style.fontWeight = "bold"
    } else {
        bold.classList.remove("target")
        icc.style.fontWeight = "normal"
    }
    params.set("bold", newState)
    boldState = newState
    window.history.pushState(null, "", `?${params.toString()}`)
}

function italicFont() {
    newState = !italicState
    const params = new URLSearchParams(window.location.search)
    const icc = document.getElementById("icc")
    const italic = document.getElementById("font-italic")
    if (newState == true) {
        italic.classList.add("target")
        icc.style.fontStyle = "italic"
    } else {
        italic.classList.remove("target")
        icc.style.fontStyle = "normal"
    }
    params.set("italic", newState)
    italicState = newState
    window.history.pushState(null, "", `?${params.toString()}`)
}

/**
 * @param {[key: string]: string} params 
 */
function applyParams(params) {
    for (const [key, value] of Object.entries(params)) {
        if (key == "brd") {
            if (value == "4") {
                borderWidth(2)
            }
            else if (value == "8") {
                borderWidth(2)
            }
        }

        if (key == "cmt") {
            if (value == "bottom") {
                overlayComments(2)
            }
            else if (value == "top") {
                overlayComments(3)
            }
        }

        if (key == "bc") {
            borderColor(value)
        }

        if (key == "font") {
            const fontSelect = document.getElementById("font-select")
            fontSelect.value = value
            changeFont(fontSelect)
        }

        if (key == "bold") {
            if (value == "true") {
                boldFont()
            }
        }

        if (key == "italic") {
            if (value == "true") {
                italicFont()
            }
        }
    }
}