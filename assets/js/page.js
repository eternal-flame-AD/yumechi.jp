function toggleTategumi(selector) {
    selector = selector || "div.post div.content"
    let existingClass = document.querySelector(selector).classList;
    if (existingClass.contains("tategumi")) {
        if (localStorage)
            localStorage.removeItem("content-tategumi");
        existingClass.remove("tategumi");
        return false;
    }
    if (localStorage)
        localStorage.setItem("content-tategumi", 1)
    existingClass.add("tategumi");
    return true;
}

$(function() {
    const sideNav = $('.sidenav');
    if (sideNav.sidenav)
        sideNav.sidenav();

    const dropDown = $('.dropdown-trigger');
    if (dropDown.dropdown)
        dropDown.dropdown();

    document.querySelectorAll("div.sensitive-img-unblur").forEach(el => {
        el.onclick = (e) => {
            el.parentElement.classList.add("acknowledged")
        }
    })

    if (localStorage) {
        if (localStorage.getItem("content-tategumi") == "1")
            toggleTategumi()
    }

});