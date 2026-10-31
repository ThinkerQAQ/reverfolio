"use strict";

// The toolbar action opens a full control panel in the browser side panel.
const page = "ui/index.html";
chrome.runtime.onInstalled.addListener(() => {
  if (chrome.sidePanel?.setPanelBehavior) {
    chrome.sidePanel.setPanelBehavior({openPanelOnActionClick:true}).catch(() => {});
  }
});
if (chrome.sidePanel?.setPanelBehavior) {
  chrome.sidePanel.setPanelBehavior({openPanelOnActionClick:true}).catch(() => {});
} else {
  chrome.action.onClicked.addListener(() => chrome.tabs.create({url:chrome.runtime.getURL(page)}));
}
