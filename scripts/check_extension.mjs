import assert from "node:assert/strict";
import {readFileSync,existsSync} from "node:fs";
import path from "node:path";
import {fileURLToPath} from "node:url";
import vm from "node:vm";
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"../extension");
const manifest=JSON.parse(readFileSync(path.join(root,"manifest.json"),"utf8"));
assert.equal(manifest.manifest_version,3);
assert.equal(manifest.name,"__MSG_extensionName__");
assert.equal(manifest.action.default_popup,"popup/index.html");
assert.equal(manifest.side_panel.default_path,"popup/index.html");
assert.deepEqual([...manifest.permissions].sort(),["sidePanel","storage"].sort());
assert.equal(manifest.host_permissions.length,3);
for(const host of manifest.host_permissions){
 const url=new URL(host.replace("*",""));
 assert.equal(url.hostname,"127.0.0.1");
 assert.equal(url.protocol,"http:");
}
const declared=new Set([...Object.values(manifest.action.default_icon),...Object.values(manifest.icons)]);
for(const item of declared){
 assert.ok(existsSync(path.join(root,item)),"Missing icon: "+item);
 const buf=readFileSync(path.join(root,item));
 assert.equal(buf.subarray(0,8).toString("hex"),"89504e470d0a1a0a");
 const size=Number(item.match(/icon(\d+)\.png/)[1]);
 assert.equal(buf.readUInt32BE(16),size);
 assert.equal(buf.readUInt32BE(20),size);
}
for(const lang of ["en","zh_CN"]){
 const messages=JSON.parse(readFileSync(path.join(root,"_locales",lang,"messages.json"),"utf8"));
 for(const key of ["extensionName","extensionDescription","actionTitle"])assert.ok(messages[key]?.message);
}
const html=readFileSync(path.join(root,"popup/index.html"),"utf8");
const script=readFileSync(path.join(root,"popup/popup.js"),"utf8");
assert.match(html,/<script src="\.\/popup\.js" defer><\/script>/);
assert.ok(!/\son\w+\s*=/.test(html),"No inline event handlers allowed");
assert.ok(!/https:\/\/|document\.cookie|nativeMessaging/.test(script),"No third-party fetching or cookie access");
const els=new Map();
class Element{
 constructor(key){
  this.key=key;this.dataset={};this.listeners={};this.textContent="";this.value="";
  this.disabled=false;this.className="dot";
 }
 addEventListener(name,handler){this.listeners[name]=handler}
 async click(){if(this.listeners.click)return await this.listeners.click({target:this})}
 async change(value){this.value=value;if(this.listeners.change)return await this.listeners.change({target:this})}
}
const get=id=>{
 if(!els.has(id))els.set(id,new Element(id));
 return els.get(id);
};
const translationElems=["tagline","retry","service","auto","windows","wsl","portable","open","quick","overview","works","channels","evidence","tasks","history","settings","sidePanel","hint","privacy"].map(key=>{
 const el=new Element(key);el.dataset.i18n=key;return el;
});
const routeElems=["overview","works","channels","evidence","tasks","history","settings"].map(route=>{
 const el=new Element("route:"+route);el.dataset.route=route;return el;
});
const writes=[],openTabs=[],healthChecks=[];
const chrome={
 storage:{local:{
  get:async()=>({language:"zh",service:"auto"}),
  set:async value=>{writes.push(value)}
 }},
 tabs:{create:async payload=>{openTabs.push(payload)}},
 sidePanel:{open:async()=>{}},
 windows:{WINDOW_ID_CURRENT:-2,getCurrent:async()=>({id:1})}
};
let closed=0;
const document={
 documentElement:{lang:""},
 getElementById:get,
 querySelectorAll:(query)=>{
  if(query==="[data-i18n]")return translationElems;
  if(query==="[data-route]")return routeElems;
  return [];
 }
};
const fetch=async(url,opts)=>{
 healthChecks.push({url,opts});
 return {
  ok:url.startsWith("http://127.0.0.1:19452/"),
  json:async()=>({status:"ok"})
 };
};
vm.runInNewContext(script,{document,chrome,fetch,window:{close:()=>closed++},setTimeout,clearTimeout,AbortController,console});
await new Promise(resolve=>setTimeout(resolve,30));
assert.equal(get("status").textContent,"本地服务已连接");
assert.equal(get("open").disabled,false);
assert.equal(get("service-address").textContent,"http://127.0.0.1:19452");
assert.ok(healthChecks.every(x=>x.opts.credentials==="omit"));
await routeElems.find(x=>x.dataset.route==="settings").click();
assert.equal(openTabs[0].url,"http://127.0.0.1:19452/#settings");
assert.equal(openTabs[0].active,true);
assert.equal(closed,1);
await get("lang").click();
assert.equal(get("lang").textContent,"中文");
assert.equal(document.documentElement.lang,"en");
assert.equal(writes[0].language,"en");
await get("service").change("wsl");
assert.equal(writes[1].service,"wsl");
assert.equal(get("status").textContent,"Local service not running");
assert.equal(get("open").disabled,true);
console.log("Extension manifest, icon sizes, locale, isolated health check and shortcut interactions passed.");
