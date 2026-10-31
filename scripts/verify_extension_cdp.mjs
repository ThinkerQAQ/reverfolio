import assert from "node:assert/strict";
import {writeFileSync} from "node:fs";
const extensionId=process.argv[2],port=Number(process.argv[3]||19506);
if(!/^[a-p]{32}$/.test(extensionId))throw Error("Missing extension id");
const root="http://127.0.0.1:"+port;
const page="chrome-extension://"+extensionId+"/ui/index.html#overview";
const response=await fetch(root+"/json/new?"+encodeURIComponent(page),{method:"PUT"});
if(!response.ok)throw Error("Cannot open extension UI "+response.status);
const target=await response.json();
const ws=new WebSocket(target.webSocketDebuggerUrl);
await new Promise((resolve,reject)=>{
 ws.addEventListener("open",resolve,{once:true});
 ws.addEventListener("error",reject,{once:true});
});
let serial=0;
const pending=new Map();
const httpRequests=[];
ws.addEventListener("message",event=>{
 const item=JSON.parse(event.data);
 if(item.method==="Network.requestWillBeSentExtraInfo"){
  const headers=item.params?.headers||{};
  const clean={origin:headers.Origin||headers.origin||"",mode:headers["Sec-Fetch-Mode"]||"",site:headers["Sec-Fetch-Site"]||"",requestId:item.params?.requestId};
  httpRequests.push(clean);
 }
 if(item.method==="Network.responseReceived"){
  const response=item.params?.response||{};
  if(response.url?.includes("/api/")){
   httpRequests.push({api:response.url.replace(/^.*\/api\//,"/api/"),status:response.status});
  }
 }
});
ws.addEventListener("message",e=>{
 const d=JSON.parse(e.data);
 if(!d.id)return;
 const entry=pending.get(d.id);
 if(!entry)return;
 pending.delete(d.id);
 if(d.error)entry.reject(Error(d.error.message));
 else entry.resolve(d.result);
});
function send(method,params={}){
 const id=++serial;
 return new Promise((resolve,reject)=>{
  pending.set(id,{resolve,reject});
  ws.send(JSON.stringify({id,method,params}));
  setTimeout(()=>{if(pending.delete(id))reject(Error("CDP timeout "+method))},12000);
 });
}
async function evaluate(expr){
 const r=await send("Runtime.evaluate",{expression:expr,awaitPromise:true,returnByValue:true});
 if(r.exceptionDetails)throw Error("Browser JS failure: "+JSON.stringify(r.exceptionDetails));
 return r.result.value;
}
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
await send("Page.enable");
await send("Network.enable");
await send("Runtime.enable");
await send("Page.reload",{ignoreCache:true});
await send("Emulation.setDeviceMetricsOverride",{width:420,height:950,deviceScaleFactor:1,mobile:false});
let complete=false,first;
for(let i=0;i<30;i++){
 await sleep(330);
 first=await evaluate("({ready:document.readyState,error:document.getElementById('notice')?.textContent,stats:document.querySelectorAll('.stat-value').length,title:document.title})");
 if(first.stats===4){complete=true;break}
}
if(!complete)console.log("NETWORK_DIAGNOSTICS",JSON.stringify(httpRequests.slice(0,28)));
assert.ok(complete,"Extension did not load metrics: "+JSON.stringify(first));
const initial=await evaluate([
 "({",
 " title:document.title,",
 " stats:[...document.querySelectorAll('.stat-card .stat-value')].map(e=>e.textContent.trim()),",
 " cards:document.querySelectorAll('.activity-card').length,",
 " sidebarWidth:Math.round(document.querySelector('.sidebar').getBoundingClientRect().width),",
 " viewport:document.documentElement.clientWidth,",
 " overflow:document.documentElement.scrollWidth-document.documentElement.clientWidth,",
 " hasServicePicker:!!document.getElementById('service'),",
 " language:document.documentElement.lang,",
 " api:apiOrigin,",
 " navigation:document.querySelectorAll('#navigation button').length",
 "})"
].join(""));
console.log("NARROW",JSON.stringify(initial));
assert.equal(initial.stats[0],"31");
assert.ok(initial.cards>=8);
assert.equal(initial.sidebarWidth,58);
assert.ok(initial.overflow<=2);
assert.equal(initial.hasServicePicker,false);
assert.equal(initial.api,"http://127.0.0.1:19452");
assert.equal(initial.navigation,7);
await evaluate("document.getElementById('locale-button').click();true");
await sleep(850);
const language=await evaluate("({lang:document.documentElement.lang,status:document.getElementById('notice').textContent})");
console.log("LANGUAGE",JSON.stringify(language));
assert.equal(language.lang,"en");
await evaluate("document.getElementById('locale-button').click();true");
await sleep(850);
assert.equal(await evaluate("document.documentElement.lang"),"zh-CN");
await evaluate("document.querySelector('[data-action=navigate][data-view=settings]').click();true");
await sleep(160);
const settings=await evaluate("({title:document.querySelector('#page-title').textContent,categories:document.querySelectorAll('#settings-home [data-action=settings-select]').length})");
console.log("SETTINGS",JSON.stringify(settings));
assert.ok(settings.categories>=6);
await evaluate("document.querySelector('#settings-home [data-id=interface]').click();true");
const depth=await evaluate("({name:document.querySelector('#settings-detail h2')?.textContent,children:document.querySelectorAll('#settings-detail [data-action=settings-select]').length})");
console.log("CATEGORY",JSON.stringify(depth));
assert.ok(depth.children>=2);
await send("Emulation.setDeviceMetricsOverride",{width:1380,height:900,deviceScaleFactor:1,mobile:false});
await sleep(100);
await evaluate("document.querySelector('[data-action=settings-home]').click();true");
const wide=await evaluate("({width:document.documentElement.clientWidth,panelWidth:Math.round(document.querySelector('#settings-home').getBoundingClientRect().width),ghost:!!document.querySelector('.settings-detail-empty')})");
console.log("WIDE",JSON.stringify(wide));
assert.ok(wide.panelWidth>1100);
assert.equal(wide.ghost,false);
await send("Emulation.setDeviceMetricsOverride",{width:420,height:950,deviceScaleFactor:1,mobile:false});
await sleep(100);
await evaluate("document.querySelector('[data-action=navigate][data-view=overview]').click();true");
await sleep(100);
const screen=await send("Page.captureScreenshot",{format:"png",captureBeyondViewport:false});
if(process.argv[4])writeFileSync(process.argv[4],Buffer.from(screen.data,"base64"));
ws.close();
console.log("REAL_EXTENSION_CDP_ACCEPTANCE_PASS");
