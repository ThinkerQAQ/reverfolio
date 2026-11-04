import assert from "node:assert/strict";
import {readFileSync,existsSync} from "node:fs";
import path from "node:path";
import {fileURLToPath} from "node:url";
import vm from "node:vm";
import crypto from "node:crypto";

const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"../");
const ext=path.join(root,"extension");
const m=JSON.parse(readFileSync(path.join(ext,"manifest.json"),"utf8"));
const read=(dir,file)=>readFileSync(path.join(dir,file),"utf8");
assert.equal(m.manifest_version,3);
assert.equal(m.name,"__MSG_extensionName__");
assert.equal(m.version,"0.11.0");
assert.equal(m.short_name,"Reverfolio");
assert.equal(m.action.default_popup,undefined,"No second launcher popup");
assert.equal(m.side_panel.default_path,"ui/index.html");
assert.equal(m.background.service_worker,"background.js");
assert.deepEqual(m.permissions,["sidePanel","nativeMessaging"]);
assert.deepEqual(m.host_permissions,["http://127.0.0.1:19452/*"]);
assert.match(m.content_security_policy.extension_pages,/connect-src 'self' http:\/\/127.0.0.1:19452/);
const publicKey=Buffer.from(m.key,"base64");
const hash=crypto.createHash("sha256").update(publicKey).digest().subarray(0,16);
const id=Array.from(hash).map(b=>String.fromCharCode(97+(b>>4),97+(b&15))).join("");
const go=read(root,"internal/echo/extension_origin.go");
assert.ok(go.includes("chrome-extension://"+id),"Backend origin must match extension identity");
const icons=new Set([...Object.values(m.action.default_icon),...Object.values(m.icons)]);
for(const item of icons){
 assert.ok(existsSync(path.join(ext,item)),item);
 const buf=readFileSync(path.join(ext,item));
 assert.equal(buf.subarray(0,8).toString("hex"),"89504e470d0a1a0a");
 const size=Number(item.match(/icon(\d+)\.png/)[1]);
 assert.equal(buf.readUInt32BE(16),size);
 assert.equal(buf.readUInt32BE(20),size);
}
for(const locale of ["zh_CN","en"]){
 const messages=JSON.parse(read(path.join(ext,"_locales",locale),"messages.json"));
 for(const key of ["extensionName","extensionDescription","actionTitle"])assert.ok(messages[key]?.message);
}
const ui=path.join(ext,"ui");
const source=path.join(root,"internal/echo/web");
for(const name of ["app.js","settings.js","dashboard.js","app.css","favicon.svg"]){
 assert.equal(read(ui,name),read(source,name),"Extension UI must mirror main UI: "+name);
}
const html=read(ui,"index.html");
assert.match(html,/<meta name="reverfolio-api-origin" content="http:\/\/127.0.0.1:19452">/);
assert.match(html,/src="\.\/app.js"/);
assert.match(html,/src="\.\/local-auth.js"/);
const template=read(ui,"local-auth.js");
assert.match(template,/REVERFOLIO_EXTENSION_ACCESS = ""/,"Repository must contain an empty credential template");
assert.match(html,/src="\.\/settings.js"/);
assert.match(html,/src="\.\/favicon.svg"/,"Brand image must load from extension package");
assert.ok(!html.includes('/assets/favicon.svg'),"No web-app absolute paths inside extension");
assert.ok(!html.includes('id="service"'),"No redundant service selector");
assert.ok(!html.includes("http://127.0.0.1:19450"),"No WSL fallback");
const app=read(ui,"app.js");
assert.match(app,/extensionMode/);
assert.match(app,/apiURL\(url\)/);
assert.match(app,/credentials:extensionMode\?"omit":"same-origin"/);
assert.match(app,/X-Reverfolio-Extension-Access/);
assert.match(app,/function offlineView\(\)/,"Offline state must provide recovery controls");
assert.match(app,/data-action="start-local-service"/);
assert.match(app,/await ensureLocalService\(\)/,"Dashboard must ensure startup before fetching data");
assert.match(app,/nativeReady=true/,"Do not start a second service after successful probe");
assert.ok(!/eval\(|new Function\(/.test(app),"Avoid unsafe-eval");
const background=read(ext,"background.js");
let enabled=0,opened=0;
let received,launcher=null;
const chrome={
 sidePanel:{setPanelBehavior:async o=>{assert.equal(o.openPanelOnActionClick,true);enabled++}},
 runtime:{
  id:"khdciakflcmilakmaiejamcgmggnkego",
  onInstalled:{addListener:()=>{}},
  onMessage:{addListener:fn=>{received=fn;}},
  sendNativeMessage:(host,request,reply)=>{
   assert.equal(host,"com.reverfolio.bridge");
   assert.equal(request.command,"ensure_bridge");
   launcher++;
   reply({ok:true,state:"running"});
  }
 },
 action:{onClicked:{addListener:()=>{opened++}}}
};
vm.runInNewContext(background,{chrome});
assert.equal(enabled,1,"Browser action must open the full side panel");
assert.equal(opened,0,"No obsolete launcher tab for a supported browser");
let nativeResponse;
assert.equal(received({type:"reverfolio.bridge",command:"ensure_bridge"},{id:chrome.runtime.id},r=>{nativeResponse=r}),true);
assert.equal(nativeResponse?.state,"running");
assert.equal(launcher,1,"Extension must request native service on demand");
assert.equal(received({type:"reverfolio.bridge",command:"ensure_bridge"},{id:"other-extension"},()=>{}),false,"Other extensions cannot trigger service startup");
console.log("Extension verified: shared dashboard, pinned origin",id,"native on-demand startup, offline manual recovery and side-panel action");
