import {readStorage, strRepeat, writeStorage} from "../utils.js";

const popupBtn = document.getElementById('my_ext-popup_btn');
popupBtn.addEventListener('click', launchInAppMode);

async function launchInAppMode() {
    await chrome.tabs.query({active: true, currentWindow: true}, (tabs) => {
        if (tabs.length === 0) {
            return;
        }
        const activeTabUrl = tabs[0].url;
        callLauncher(activeTabUrl);
    });
}

async function callLauncher(url) {
    const encodedUrl = 'http://localhost:5252/open/' + url
        .replace('://', '~')
        .replaceAll('/', '+');

    await fetch(encodedUrl);
}