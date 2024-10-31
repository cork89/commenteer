// var enc = new TextEncoder()
// var dec = new TextDecoder()

function validateLink() {
    let url = document.getElementById("link");

    let _, error = isValidRedditUrl(url.value)
    if (error) {
        document.getElementById("linkError").innerText = error;
    } else {
        let [subreddit, article, comment] = getLinkTokens(url.value)
        window.location.href = `/r/${subreddit}-${article}-${comment}/submit`
    }
}

function isValidRedditUrl(input) {
    let url;
    try {
        url = new URL(input);
        if (url.host !== "reddit.com" && url.host !== "www.reddit.com") {
            return false, "Not a reddit link";
        }
    } catch (_) {
        return false, "Not a valid url";
    }
    return url.protocol === "http:" || url.protocol === "https:", "";
}

function getLinkTokens(link) {
    let tokens = link.split("/")
    let [https, url, r, subreddit, comments, article, title, comment] = tokens.filter(item => item != '')
    return [subreddit, article, comment]
}

async function copyImage(event) {
    try {
        const queryId = event.id.substr(0, event.id.indexOf("-copy"))
        const imgDiv = document.getElementById(queryId)
        const img = imgDiv.src

        // let imgData = sessionStorage.getItem(imgId);
        // if (!imgBlob) {
        const imgData = await fetch(`http://localhost:8090/image/?src=${img}`, {
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
    } catch (error) {
        console.error(error);
    }
}
