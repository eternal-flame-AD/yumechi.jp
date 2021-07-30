function toggleTategumi(selector) {
    selector = selector || "div.content"
    let existingClass = document.querySelector(selector).classList;
    if (existingClass.contains("tategumi")) {
        existingClass.remove("tategumi");
        return false;
    }
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
});