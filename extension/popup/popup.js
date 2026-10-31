"use strict";

// A launcher and health viewer only: it never receives GitHub, MCP, or SQLite secrets.
const SERVICES = Object.freeze({
  windows: "http://127.0.0.1:19452",
  wsl: "http://127.0.0.1:19450",
  portable: "http://127.0.0.1:18440"
});
const TEXT = {
  zh:{
    tagline:"影响力观察工作台",retry:"刷新",service:"本地服务",
    auto:"自动检测（优先 Windows）",windows:"Windows · 19452",wsl:"Ubuntu / WSL · 19450",portable:"其他本地 · 18440",
    open:"打开 OpenEcho",quick:"快捷入口",overview:"影响力总览",works:"项目管理",channels:"分发渠道",
    evidence:"引用与采用",tasks:"发现任务",history:"历史统计",settings:"设置",
    sidePanel:"在侧边栏打开快捷入口",
    hint:"扩展只检查本地服务状态，实际数据管理在 OpenEcho 页面完成。",
    privacy:"不读取账号 Cookie · 不上传数据",checking:"检测中…",online:"本地服务已连接",
    offline:"本地服务未启动",errorHint:"请先运行 opencho.exe，再刷新扩展。"
  },
  en:{
    tagline:"Impact observatory",retry:"Refresh",service:"Local service",
    auto:"Auto-detect (Windows first)",windows:"Windows · 19452",wsl:"Ubuntu / WSL · 19450",portable:"Other local · 18440",
    open:"Open OpenEcho",quick:"Quick access",overview:"Impact overview",works:"Projects",channels:"Channels",
    evidence:"Evidence & adoption",tasks:"Discovery tasks",history:"History",settings:"Settings",
    sidePanel:"Open shortcuts in side panel",
    hint:"This extension only checks local service health. Manage data in the OpenEcho dashboard.",
    privacy:"No account cookies · No uploads",checking:"Checking…",online:"Local service connected",
    offline:"Local service not running",errorHint:"Start opencho.exe and refresh this extension."
  }
};
const byId=id=>document.getElementById(id);
const app={language:"zh",choice:"auto",active:null,checking:false};
const translation=key=>TEXT[app.language][key]||key;
const allowableRoutes=new Set(["overview","works","channels","evidence","tasks","history","settings"]);
function renderLabels(){
  document.documentElement.lang=app.language==="en"?"en":"zh-CN";
  document.querySelectorAll("[data-i18n]").forEach(el=>{el.textContent=translation(el.dataset.i18n)});
  byId("lang").textContent=app.language==="zh"?"EN":"中文";
  byId("service").value=app.choice;
  renderStatus();
}
function renderStatus(){
  const status=byId("status"),dot=byId("status-dot");
  status.textContent=app.checking?translation("checking"):app.active?translation("online"):translation("offline");
  dot.className="dot"+(!app.checking?(app.active?" online":" offline"):"");
  byId("service-address").textContent=app.active||(
    app.choice==="auto"?"127.0.0.1 · 19452 / 19450 / 18440":SERVICES[app.choice]);
  for(const item of [byId("open"),...document.querySelectorAll("[data-route]")])item.disabled=!app.active;
  byId("help").textContent=app.active?translation("hint"):translation("errorHint");
}
async function live(origin){
 const controller=new AbortController();
 const timeout=setTimeout(()=>controller.abort(),1200);
 try{
  const response=await fetch(origin+"/healthz",{method:"GET",credentials:"omit",cache:"no-store",signal:controller.signal});
  if(!response.ok)return false;
  const json=await response.json();
  return json.status==="ok";
 }catch{return false}
 finally{clearTimeout(timeout)}
}
async function detect(){
 if(app.checking)return;
 app.checking=true;app.active=null;renderStatus();
 const list=app.choice==="auto"?Object.keys(SERVICES):[app.choice];
 for(const name of list){
  const origin=SERVICES[name];
  if(await live(origin)){app.active=origin;break}
 }
 app.checking=false;renderStatus();
}
async function openRoute(route){
 if(!app.active||!allowableRoutes.has(route))return;
 await chrome.tabs.create({url:app.active+"/#"+route,active:true});
 window.close();
}
async function initialize(){
 try{
  const stored=await chrome.storage.local.get(["language","service"]);
  if(stored.language==="zh"||stored.language==="en")app.language=stored.language;
  if(stored.service==="auto"||stored.service in SERVICES)app.choice=stored.service;
 }catch{}
 renderLabels();
 await detect();
}
byId("lang").addEventListener("click",async()=>{
 app.language=app.language==="zh"?"en":"zh";
 await chrome.storage.local.set({language:app.language});
 renderLabels();
});
byId("service").addEventListener("change",async e=>{
 const selected=e.target.value;
 if(selected!=="auto"&&!(selected in SERVICES))return;
 app.choice=selected;
 await chrome.storage.local.set({service:app.choice});
 await detect();
});
byId("retry").addEventListener("click",detect);
byId("open").addEventListener("click",()=>openRoute("overview"));
document.querySelectorAll("[data-route]").forEach(button=>button.addEventListener("click",()=>openRoute(button.dataset.route)));
byId("sidepanel").addEventListener("click",async()=>{
 try{
  if(!chrome.sidePanel?.open)throw Error("unsupported");
  await chrome.sidePanel.open({windowId:chrome.windows.WINDOW_ID_CURRENT});
  window.close();
 }catch{
  if(app.active)await openRoute("overview");
 }
});
initialize();
