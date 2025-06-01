chrome.runtime.onInstalled.addListener(({reason}) => {
    if (reason === 'install') {
      chrome.tabs.create({
        url: "src/onboarding/onboarding.html"
      });
    }
  });


chrome.action.onClicked.addListener(async (tab) => {
    // if (!await statusCheck()) {
    //     window.alert("CalGo server is not running!");
    //     return;
    // }
    await requestAppLaunch(tab.url)
});


async function statusCheck() {
    return await fetch('http://localhost:5252/status')
        .then(response => {
            return response.ok;
        });
}

async function requestAppLaunch(url) {
    const encodedUrl = 'http://localhost:5252/open/' + url
        .replace('://', '~')
        .replaceAll('/', '+');

    await fetch(encodedUrl);
}