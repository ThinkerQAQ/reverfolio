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


// The extension uses the registered native host solely for checking/starting
// its single Windows loopback service. Never put bearer tokens in messages.
const REVERFOLIO_NATIVE_HOST = "com.reverfolio.bridge";
chrome.runtime.onMessage.addListener((message,sender,reply)=>{
  if(sender.id!==chrome.runtime.id || message?.type!=="reverfolio.bridge")return false;
  const command=message.command==="status"?"status":message.command==="ensure_bridge"?"ensure_bridge":null;
  if(!command){reply({ok:false,error:"unsupported command"});return false;}
  chrome.runtime.sendNativeMessage(REVERFOLIO_NATIVE_HOST,{command},response=>{
    const error=chrome.runtime.lastError;
    reply(error?{ok:false,error:"native launcher unavailable: "+error.message}:(response||{ok:false,error:"native launcher returned no response"}));
  });
  return true;
});
