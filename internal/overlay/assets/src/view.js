const updateElem = (elem) => {
    if (elem.dataset.left != "0") {
        elem.style.left = elem.dataset.left + "px";
    } else {
        elem.style.left = null;
    }
    if (elem.dataset.right != "0") {
        elem.style.right = elem.dataset.right + "px";
    } else {
        elem.style.right = null;
    }
    if (elem.dataset.top != "0") {
        elem.style.top = elem.dataset.top + "px";
    } else {
        elem.style.top = null;
    }
    if (elem.dataset.bottom != "0") {
        elem.style.bottom = elem.dataset.bottom + "px";
    } else {
        elem.style.bottom = null;
    }
    if (elem.dataset.bgcolor != "") {
        elem.style.backgroundColor = elem.dataset.bgcolor;
    } else {
        elem.style.backgroundColor = null;
    }
};
const onMutation = (mutationsList) => {
    mutationsList.forEach((mutation) => {
        updateElem(mutation.target);
    });
};
const view = () => {
    const observer = new MutationObserver(onMutation);
    Array.from(document.querySelectorAll('*[data-left]'), (elem) => {
        updateElem(elem);
        observer.observe(elem, { attributes: true });
    });
};
view();