var offset = 1
var isFetching = false
const maincontent = document.getElementById("maincontent")
const page = document.getElementsByName("page")[0].content
const subpage = document.getElementsByName("subpage")[0].content

async function handleScrollPosition() {
    if (isFetching) {
        return
    }
    if (offset > 10) {
        window.removeEventListener("scroll", handleScrollPosition)
        return
    }
    const scrollTop = window.PageYOffset || this.document.documentElement.scrollTop
    if ((scrollTop / (maincontent.scrollHeight - window.innerHeight)) > 0.75) {
        offset += 1
        try {
            isFetching = true
            const response = await fetch(`/data/?offset=${offset}&page=${page}&subpage=${subpage}`, { cache: "no-store" })
            const data = await response.text()
            maincontent.insertAdjacentHTML("beforeend", data)
        } catch (error) {
            console.error("error fetching data: ", error)
        } finally {
            isFetching = false
        }
    }
}

window.addEventListener("scroll", handleScrollPosition)