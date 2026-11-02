"use strict";

/*
 * Recursive standalone settings navigator.  Presentation pattern inspired by
 * familiar desktop setting panels, implemented independently for Reverfolio.
 * Leaf controls are bound only to Reverfolio's allowlisted settings API.
 */
const settingTree = [
 {id:"mcp",name:["MCP 连接","MCP connections"],desc:["本地客户端和 ChatGPT Tunnel","Local clients and ChatGPT Tunnel"],children:[{id:"mcp-connection",name:["连接与服务管理","Connections and service"],desc:["配置、健康检查、启动与停止","Configuration, readiness, start and stop"],kind:"mcp-connection"}]},
 {id:"interface",name:["界面与语言","Interface & language"],desc:["界面语言、首页、布局","Language, dashboard and layout"],children:[
   {id:"ui-language",name:["界面语言","Interface language"],desc:["简体中文或 English（保存在 config.toml）","Simplified Chinese or English (saved to config.toml)"],kind:"language"},
   {id:"dashboard",name:["首页总览","Dashboard"],desc:["调整首页的展示方式","Control the overview layout"],children:[
     {id:"dashboard-sections",name:["显示模块","Visible sections"],desc:["GitHub、引用证据、历史统计","GitHub, evidence and historical metrics"],fields:[
       {key:"overview.show_github",name:["GitHub 项目活动","GitHub activity"],help:["首页是否显示 GitHub 指标表格","Show GitHub metrics table on the overview"],type:"boolean"},
       {key:"overview.show_evidence",name:["外部引用证据","External evidence"],help:["首页是否展示经核实的引用","Show verified citations and adoption"],type:"boolean"},
       {key:"overview.show_history",name:["历史数据摘要","Historical summary"],help:["首页是否展示历史 GSC、Umami 指标","Show historical Search Console and Umami observations"],type:"boolean"}
     ]},
     {id:"dashboard-volume",name:["证据显示数量","Evidence display limit"],desc:["首页最近证据数量","Number of recent evidence items"],fields:[
       {key:"overview.evidence_limit",name:["最近证据条数","Recent evidence items"],help:["仅影响首页展示数量，不影响数据库","Only limits display, not database records"],type:"select",options:[["2","2"],["4","4"],["8","8"],["12","12"]]}
     ]}
   ]},
   {id:"ui-appearance",name:["布局密度","Layout density"],desc:["调整管理工作台的间距","Adjust workspace spacing"],fields:[
     {key:"appearance.compact",name:["紧凑布局","Compact layout"],help:["减少表格和卡片间距","Reduce table and card padding"],type:"boolean"}
   ]}
 ]},
 {id:"data-sources",name:["数据采集","Data sources"],desc:["GitHub、搜索平台、访问统计","GitHub, search and analytics"],children:[
   {id:"github",name:["GitHub","GitHub"],desc:["下载口径及采集运行说明","Download scope and collector status"],children:[
     {id:"github-downloads",name:["Release 下载统计","Release downloads"],desc:["切换全部资产或可安装程序包","Count all assets or installable packages"],fields:[
       {key:"github.download_scope",name:["统计口径","Download scope"],help:["「安装包」仅筛选常见二进制发布包，仍无法去除机器人或重复下载","Installables exclude common checksums/sources, but cannot identify real users"],type:"select",options:[["all",["全部 Release 资产","All release assets"]],["installers",["安装包（过滤辅助文件）","Installables (exclude auxiliary files)"]]]}
     ]},
     {id:"github-schedule",name:["自动采集计划","Automatic collection"],desc:["设置采集间隔，无需重启","Set the interval without restarting"],fields:[
       {key:"collection.interval",name:["采集频率","Collection interval"],help:["一分钟内生效；电脑和本地服务须保持运行。关闭计划不会删除已采集数据。","Applied within one minute while the computer and local service are running. Disabling preserves observations."],type:"select",options:[["0",["关闭","Disabled"]],["1h",["每小时","Hourly"]],["6h",["每 6 小时","Every 6 hours"]],["24h",["每天","Daily"]],["168h",["每周","Weekly"]]]}
     ]},
     {id:"github-auth",name:["GitHub 授权方式","GitHub authentication"],desc:["更改后重启服务生效","Restart the service after changing"],fields:[
 {key:"github.auth",name:["凭据来源","Credential source"],help:["环境变量 GITHUB_TOKEN 或本机已登录的 gh；密钥不写入配置。更改后重启。","Use GITHUB_TOKEN or the signed-in local gh CLI. Secrets are not written to config. Restart required."],type:"select",options:[["env",["环境变量","Environment"]],["gh",["本机 GitHub CLI","Local GitHub CLI"]]]}]},
 {id:"github-runtime",name:["采集运行状态","Collection runtime"],desc:["定时采集与账号授权","Scheduling and authorization"],kind:"github-runtime"}
   ]},
   {id:"search-sources",name:["搜索与站点统计","Search and site analytics"],desc:["GSC、Bing、Umami 当前支持情况","Current GSC, Bing and Umami coverage"],children:[
     {id:"search-gsc",name:["Google Search Console","Google Search Console"],desc:["搜索曝光和点击","Impressions and clicks"],kind:"gsc"},
     {id:"search-bing",name:["Bing Webmaster","Bing Webmaster"],desc:["Bing 搜索统计","Bing search statistics"],kind:"bing"},
     {id:"search-umami",name:["Umami","Umami"],desc:["网站浏览与访问标识","Pageviews and visitor identifiers"],kind:"umami"}
   ]}
 ]},
 {id:"review",name:["引用与审核","Evidence & review"],desc:["审核人员及证据过滤","Review identity and filters"],children:[
   {id:"review-process",name:["审核流程","Review workflow"],desc:["可追溯的审核记录","Audited reviews"],children:[
     {id:"review-author",name:["默认审核人","Default reviewer"],desc:["记录到审核历史的标识","Stored in review audit records"],fields:[
       {key:"evidence.reviewer",name:["审核人名称","Reviewer name"],help:["仅用于审核审计显示，不是账号认证","Used for audit labels; not an authentication identity"],type:"text",max:80}
     ]},
     {id:"review-filter",name:["证据默认筛选","Default evidence filter"],desc:["进入证据页面时默认显示的状态","Initial filter when opening evidence"],fields:[
       {key:"evidence.default_filter",name:["默认状态","Default status"],help:["只影响展示，不会删除或改变审核状态","Display only; does not alter review states"],type:"select",options:[["all",["全部","All"]],["verified",["已审核","Verified"]],["pending",["待审核","Pending"]]]}
     ]}
   ]}
 ]},
 {id:"discovery-settings",name:["AI 检索任务","AI discovery"],desc:["搜索任务执行与记录","Search tasks and execution"],children:[
   {id:"discovery-worker",name:["默认任务认领者","Default task worker"],desc:["UI 认领任务时使用的名称","Name used when claiming a task"],fields:[
     {key:"discovery.worker",name:["任务认领名称","Worker name"],help:["记录到任务审计，不会自动执行 AI 搜索","Recorded in task audit; does not run AI searches automatically"],type:"text",max:80}
   ]},
   {id:"discovery-runtime",name:["执行方式","Execution mode"],desc:["了解 AI 搜索和 Reverfolio 的职责","What AI clients and Reverfolio do"],kind:"search-runtime"}
 ]},
 {id:"history-settings",name:["统计和备份","Statistics & backup"],desc:["时间展示及本地数据备份","Time display and private data"],children:[
   {id:"display-time",name:["日期与时区","Dates & timezone"],desc:["统计时间的显示时区","Displayed time zone"],fields:[
     {key:"display.timezone",name:["时间显示","Displayed timezone"],help:["原始记录始终保留 UTC 时间戳；仅改变界面格式","Original timestamps remain UTC; display only"],type:"select",options:[["local",["浏览器本地时间","Browser local time"]],["UTC","UTC"]]}
   ]},
   {id:"backup-location",name:["数据库备份","SQLite backup"],desc:["导出和恢复的安全说明","Export and restore guidance"],kind:"backup"}
 ]},
 {id:"system-settings",name:["运行环境","Runtime"],desc:["本地访问与独立部署","Local access and standalone deployment"],children:[
   {id:"startup-config",name:["高级启动配置","Advanced startup configuration"],desc:["监听地址、数据库；重启后生效","Listener and database; restart required"],fields:[
 {key:"server.listen",name:["本机监听地址","Loopback listener"],help:["仅允许回环 IP 与端口。保存同步扩展地址；重启服务后访问新地址，并重新加载扩展。","Loopback IP and port only. Saving synchronizes extension metadata. Restart the service, open the new address, and reload the extension."],type:"text",max:100},
 {key:"storage.database",name:["SQLite 数据库路径","SQLite database path"],help:["高级操作：仅可选择已存在的 Reverfolio SQLite 文件。相对路径以配置文件目录为准；不会移动或复制旧数据，重启后生效。","Advanced: select an existing Reverfolio SQLite file. Relative to the config directory. Does not move or copy data. Restart required."],type:"text",max:1000}]},
 {id:"local-security",name:["本地访问与安全","Local access & security"],desc:["本地网页会话与 MCP 鉴权","Browser session and MCP authentication"],kind:"security"},
   {id:"about-reverfolio",name:["关于 Reverfolio","About Reverfolio"],desc:["版本、架构及独立性","Version, architecture and independence"],kind:"about"}
 ]}
];
const settingsLabel = pair => Array.isArray(pair) ? pair[state.lang==="en"?1:0] : pair;
const settingsValue = key => (state.data?.settings||[]).find(x=>x.key===key);
function pref(key,fallback=""){
 return settingsValue(key)?.value ?? fallback;
}
function settingFind(id,nodes=settingTree,ancestors=[]){
 for(const item of nodes){
  const trail=[...ancestors,item];
  if(item.id===id)return {node:item,trail};
  if(item.children){
   const found=settingFind(id,item.children,trail);
   if(found)return found;
  }
 }
 return null;
}
function settingLeaves(nodes){
 return nodes.reduce((n,node)=>n+(node.children?settingLeaves(node.children):1),0);
}
function settingSearch(query){
 const words=query.toLowerCase().trim().split(/\s+/).filter(Boolean);
 const out=[];
 function descend(nodes,trail=[]){
  for(const node of nodes){
   const path=[...trail,node];
   if(node.children){descend(node.children,path);}
   else {
    const text=path.map(x=>settingsLabel(x.name)+" "+settingsLabel(x.desc)+" "+
      (Array.isArray(x.name)?x.name.join(" "):"")+
      (x.fields||[]).map(f=>f.key+" "+settingsLabel(f.name)+" "+settingsLabel(f.help)+" "+
        (Array.isArray(f.name)?f.name.join(" "):"")).join(" ")).join(" ").toLowerCase();
    if(words.every(w=>text.includes(w)))out.push(path);
   }
  }
 }
 descend(settingTree);return out;
}
function settingsTreeRow(node,depth=0,filterIds=null){
 if(filterIds&&!filterIds.has(node.id))return "";
 const hasChildren=!!node.children,expanded=state.settingsExpanded.has(node.id);
 const selected=state.settingsSelected===node.id;
 const count=hasChildren?settingLeaves(node.children):null;
 return '<div class="settings-tree-branch" role="treeitem" aria-expanded="'+(hasChildren?expanded:"false")+'">'+
  '<button type="button" class="settings-tree-row '+(selected?"is-selected":"")+'" data-action="settings-node" data-id="'+E(node.id)+'" '+
  'data-depth="'+depth+'"><span class="settings-tree-marker">'+(hasChildren?(expanded?"⌄":"›"):"·")+'</span>'+
  '<span class="settings-tree-copy"><strong>'+E(settingsLabel(node.name))+'</strong><small>'+E(settingsLabel(node.desc))+'</small></span>'+
  (hasChildren?'<span class="settings-tree-count">'+count+'</span>':'<span class="settings-tree-next">›</span>')+
  '</button>'+
  (hasChildren && expanded?'<div class="settings-tree-children" role="group">'+node.children.map(x=>settingsTreeRow(x,depth+1,filterIds)).join("")+'</div>':"")+
 '</div>';
}
function settingsFilteredIDs(){
 const ids=new Set();
 for(const path of settingSearch(state.settingsQuery)){
  for(const n of path)ids.add(n.id);
 }
 return ids;
}
function settingsSidebar(){
 if(state.settingsQuery){
  const hits=settingSearch(state.settingsQuery);
  return '<div class="settings-results-count">'+E(settingsLabel(["匹配","Matches"]))+' '+hits.length+' '+E(settingsLabel(["项配置","settings"]))+'</div>'+
   (hits.length?hits.map(path=>{
    const leaf=path.at(-1);
    return '<button type="button" class="settings-search-hit" data-action="settings-select" data-id="'+E(leaf.id)+'">'+
     '<strong>'+E(settingsLabel(leaf.name))+'</strong><small>'+E(path.slice(0,-1).map(x=>settingsLabel(x.name)).join(" / "))+'</small></button>';
   }).join(""):'<div class="settings-empty">'+E(settingsLabel(["没有找到匹配的配置项","No matching settings"]))+'</div>');
 }
 return settingTree.map(n=>settingsTreeRow(n)).join("");
}
function settingsFieldInput(field){
 const saved=settingsValue(field.key);
 if(!saved)return '<p class="notice error">Unavailable setting: '+E(field.key)+'</p>';
 const value=state.settingsDraft[field.key]??saved.value;
 const control=field.type==="boolean" ?
  '<select data-setting-key="'+E(field.key)+'"><option value="true" '+(value==="true"?"selected":"")+'>'+
   E(settingsLabel(["开启","On"]))+'</option><option value="false" '+(value==="false"?"selected":"")+'>'+
   E(settingsLabel(["关闭","Off"]))+'</option></select>' :
  field.type==="select"?
   '<select data-setting-key="'+E(field.key)+'">'+(!field.options.some(o=>o[0]===value)?'<option value="'+E(value)+'" selected>'+E(value)+'</option>':'')+field.options.map(opt=>
     '<option value="'+E(opt[0])+'" '+(value===opt[0]?"selected":"")+'>'+E(settingsLabel(opt[1]))+'</option>').join("")+'</select>' :
  '<input data-setting-key="'+E(field.key)+'" maxlength="'+(field.max||80)+'" value="'+E(value)+'" required>';
 return '<div class="settings-field"><label><strong>'+E(settingsLabel(field.name))+'</strong><small>'+E(settingsLabel(field.help))+'</small>'+control+'</label></div>';
}
function settingsReadOnly(node){
 const kind=node.kind;
 if(kind==="mcp-connection")return mcpSettingsPanel();
 const status=(label,badge,detail)=>'<div class="settings-status"><div class="work-header"><strong>'+E(label)+'</strong>'+tag(E(badge),badge==="Active"?"complete":"partial")+'</div>'+
  '<p class="muted">'+E(detail)+'</p></div>';
 if(kind==="language")
  return '<div class="btn-row"><button type="button" class="btn '+(state.lang==="zh"?"btn-primary":"")+'" data-action="set-locale" data-locale="zh">简体中文</button>'+
   '<button type="button" class="btn '+(state.lang==="en"?"btn-primary":"")+'" data-action="set-locale" data-locale="en">English</button></div>'+
   '<p class="kpi-note">'+E(settingsLabel(["语言保存在 config.toml 中，不改变博客文章和统计数据。","Language is stored in config.toml without changing articles or metrics."]))+'</p>';
 if(kind==="github-runtime"){
  const runtime=state.data.runtime||{};
  const active=runtime.github_collection_enabled;
  return status("GitHub",active?"Active":"Disabled",settingsLabel([
   active?"当前定时采集间隔："+runtime.collection_interval+"；凭据方式："+runtime.github_auth_mode+"。成功率需查看快照。":"当前已关闭自动采集；凭据方式："+runtime.github_auth_mode+"。可在自动采集计划中开启。",
   active?"Collection interval: "+runtime.collection_interval+"; auth mode: "+runtime.github_auth_mode+". Check snapshots for actual success.":"Automatic collection is disabled in TOML. Auth mode: "+runtime.github_auth_mode+". Enable it in Automatic collection after configuring authorization."
  ])) + '<p>'+E(settingsLabel([runtime.github_credential_loaded?"已加载 GitHub 凭据；实际权限与采集结果请查看快照。":"未加载 GitHub 凭据，流量统计可能不可用。",runtime.github_credential_loaded?"GitHub credential loaded; inspect snapshots for permissions and actual results.":"No GitHub credential loaded; traffic metrics may be unavailable."]))+'</p>';
 }
 if(kind==="gsc"||kind==="bing"||kind==="umami")return siteSourcePanel(kind);
 if(kind==="search-runtime")return '<p>'+E(settingsLabel(["Reverfolio 保存检索任务、任务认领和证据审核；实际搜索由已授权的 AI 客户端执行，系统不会自动声称搜索了全网。","Reverfolio stores tasks, leases and evidence review. Actual search happens through an authorized AI client; no fake autonomous crawl."]))+'</p>';
 if(kind==="backup")return '<p class="desc">'+E(settingsLabel(["数据库备份是私有的 SQLite 文件。恢复需要在服务离线时使用新的文件路径，不会覆盖现有数据库。","Backups are private SQLite files. Restore only to a new destination while offline."]))+'</p>'+
  '<button type="button" class="btn btn-primary" data-action="backup">'+E(settingsLabel(["下载数据库备份","Download SQLite backup"]))+'</button>';
 if(kind==="security")return status("Local UI","Active",settingsLabel(["浏览器通过 HttpOnly 本地会话使用，MCP 和程序 API 保持 Bearer 授权，服务监听本地回环地址。","Browser uses an HttpOnly local session; MCP and API require Bearer authentication on a loopback listener."]))+
  '<div class="settings-info"><code>'+E(apiOrigin)+'</code></div>';
 if(kind==="about")return '<p><strong>Reverfolio · '+E(state.data.runtime?.version||'unknown')+'</strong></p><p>'+E(settingsLabel(["Go + SQLite + 原生嵌入式 Web UI + MCP。没有 BlogCTL 运行依赖；界面设计借鉴分层分类交互。","Go + SQLite + embedded Web UI + MCP. No BlogCTL dependency; only hierarchical navigation is an interaction reference."]))+'</p>';
 return '<div class="settings-empty">'+E(settingsLabel(["请选择一项设置","Choose a setting"]))+'</div>';
}
function settingsDetail(){
 if(!state.settingsSelected){
  return '<div class="settings-detail-empty"><div class="settings-detail-symbol">⚙</div><h2>'+E(settingsLabel(["按分类查找配置","Explore settings by category"]))+'</h2>'+
   '<p>'+E(settingsLabel(["展开左侧分类并选择设置项。配置支持搜索、编辑、保存和取消。","Expand categories and choose a setting. Search, edit, save and cancel are supported."]))+'</p></div>';
 }
 const found=settingFind(state.settingsSelected);
 if(!found)return '<div class="settings-empty">Setting not found</div>';
 const {node,trail}=found;
 const breadcrumbs='<button type="button" class="settings-crumb settings-crumb-home" data-action="settings-home">'+E(settingsLabel(["设置","Settings"]))+'</button>'+
 '<span class="settings-crumb-sep">/</span>'+trail.map((p,i)=>'<button type="button" class="settings-crumb" data-action="settings-select" data-id="'+E(p.id)+'">'+E(settingsLabel(p.name))+'</button>').join('<span class="settings-crumb-sep">/</span>');
 if(node.children)return '<div class="settings-breadcrumb">'+breadcrumbs+'</div><h2>'+E(settingsLabel(node.name))+'</h2>'+
  '<p class="desc">'+E(settingsLabel(node.desc))+'</p>'+
  '<div class="settings-category-list">'+node.children.map(child=>
   '<button type="button" class="settings-category-card" data-action="settings-select" data-id="'+E(child.id)+'"><span><strong>'+E(settingsLabel(child.name))+'</strong>'+
    '<small>'+E(settingsLabel(child.desc))+'</small></span><span>›</span></button>').join("")+'</div>';
 const controls=node.fields?
  '<form data-form="setting-edit"><div class="settings-fields">'+node.fields.map(settingsFieldInput).join("")+'</div>'+
  '<div class="settings-savebar"><span class="settings-hint">'+E(settingsLabel(["修改在保存后生效","Changes take effect after saving"]))+'</span>'+
  '<button class="btn" type="button" data-action="settings-cancel">'+E(settingsLabel(["取消","Cancel"]))+'</button>'+
  '<button class="btn btn-primary" type="button" data-action="settings-save">'+E(settingsLabel(["保存设置","Save settings"]))+'</button></div></form>':
  settingsReadOnly(node);
 return '<div class="settings-breadcrumb">'+breadcrumbs+'</div><h2>'+E(settingsLabel(node.name))+'</h2>'+
  '<p class="desc">'+E(settingsLabel(node.desc))+'</p>'+
  '<hr class="separator">'+controls;
}
// A single full-width settings route, like BlogCTL: index -> category -> leaf.
function settingsIndexRows(){
 if(state.settingsQuery){
  const hits=settingSearch(state.settingsQuery);
  return '<div class="settings-results-count">'+E(settingsLabel(["匹配","Matches"]))+' '+hits.length+' '+E(settingsLabel(["项配置","settings"]))+'</div>'+
   (hits.length?hits.map(path=>{
    const leaf=path.at(-1);
    return '<button type="button" class="settings-category-card settings-search-hit" data-action="settings-select" data-id="'+E(leaf.id)+'">'+
     '<span><strong>'+E(settingsLabel(leaf.name))+'</strong>'+
     '<small>'+E(path.slice(0,-1).map(x=>settingsLabel(x.name)).join(" / "))+'</small></span><span class="settings-category-chevron">›</span></button>';
   }).join(""):'<div class="settings-empty">'+E(settingsLabel(["没有找到匹配的配置项","No matching settings"]))+'</div>');
 }
 return settingTree.map(node=>
  '<button type="button" class="settings-category-card" data-action="settings-select" data-id="'+E(node.id)+'">'+
  '<span><strong>'+E(settingsLabel(node.name))+'</strong><small>'+E(settingsLabel(node.desc))+'</small></span>'+
  '<span class="settings-category-meta">'+settingLeaves(node.children||[])+' <span class="settings-category-chevron">›</span></span></button>'
 ).join("");
}
function renderSettingsTree(){
 if(state.settingsSelected){
  return '<section class="settings-full-page panel" id="settings-detail">'+settingsDetail()+'</section>';
 }
 return '<section class="settings-full-page panel" id="settings-home">'+
  '<div class="settings-browser-header"><h2>'+E(settingsLabel(["设置","Settings"]))+'</h2>'+
  '<p>'+E(settingsLabel(["按分类查找设置，选择后进入下一层，支持编辑、保存和返回。","Browse categories, drill down, edit, save and go back."]))+'</p></div>'+
  '<div class="settings-search-wrap"><input type="search" id="settings-search" autocomplete="off" placeholder="'+E(settingsLabel(["搜索设置：GitHub、审核、时区…","Search: GitHub, review, timezone…"]))+'" aria-label="'+E(settingsLabel(["搜索设置","Search settings"]))+'" value="'+E(state.settingsQuery)+'"></div>'+
  '<div id="settings-tree" class="settings-category-list" role="navigation" aria-label="'+E(settingsLabel(["设置分类","Settings categories"]))+'">'+settingsIndexRows()+'</div></section>';
}
function settingsRenderSidebar(){
 const area=$("settings-tree");if(area)area.innerHTML=settingsIndexRows();
}
async function settingsHandleAction(button){
 if(button.dataset.action.startsWith("settings-mcp-")){
  const action=button.dataset.action.slice("settings-mcp-".length);
  if(action==="copy"){await navigator.clipboard.writeText(JSON.stringify(state.mcpConnection?.local_config||{},null,2));notification(settingsLabel(["已复制","Copied"]));return;}
  if(action!=="refresh"){
   const tunnel_id=document.querySelector("#mcp-tunnel-id")?.value||"";
   await act(async()=>{await api("/api/mcp-connection",{method:"POST",body:JSON.stringify({action,tunnel_id})});},settingsLabel(["操作已提交，请刷新检查状态；首次配置请在本机终端完成。","Action submitted. Refresh status; finish first setup in the local terminal."]));
  }
  try{state.mcpConnection=await api("/api/mcp-connection");render();}catch(err){notification(err.message,true);}return;
 }
 if(button.dataset.id==="mcp-connection"){try{state.mcpConnection=await api("/api/mcp-connection");}catch(err){notification(err.message,true);}}

 switch(button.dataset.action){
  case "settings-home":{
   state.settingsSelected="";
   state.settingsQuery="";
   state.settingsDraft={};
   render();break;
  }
  case "settings-save":{
   await settingsHandleForm();
   break;
  }
  case "settings-node":{
   const id=button.dataset.id,found=settingFind(id);if(!found)return;
   if(found.node.children){
    if(state.settingsExpanded.has(id))state.settingsExpanded.delete(id);
    else state.settingsExpanded.add(id);
   }
   state.settingsSelected=id;render();break;
  }
  case "settings-select":{
   const id=button.dataset.id,found=settingFind(id);if(!found)return;
   for(const n of found.trail.slice(0,-1))state.settingsExpanded.add(n.id);
   state.settingsSelected=id;state.settingsQuery="";render();break;
  }
  case "settings-cancel":{
   const found=settingFind(state.settingsSelected);
   for(const f of found?.node.fields||[])delete state.settingsDraft[f.key];
   render();break;
  }
 }
}
function settingsHandleInput(input){
 if(input.id==="settings-search"){state.settingsQuery=input.value;settingsRenderSidebar();return;}
 if(input.dataset.settingKey){state.settingsDraft[input.dataset.settingKey]=input.value;}
}
async function settingsHandleForm(){
 const found=settingFind(state.settingsSelected),fields=found?.node.fields||[];
 if(!fields.length)return;
 const updates=fields.map(field=>{
  const row=settingsValue(field.key);
  const input=document.querySelector('[data-setting-key="'+field.key+'"]');
  return {key:field.key,value:input?.value??row.value,expected_version:row.version};
 }).filter(item=>item.value!==settingsValue(item.key)?.value);
 if(!updates.length){notification(settingsLabel(["没有需要保存的更改","No changes to save"]));return;}
 await act(async()=>{
  await send("update_settings",{updates});
  for(const item of updates)delete state.settingsDraft[item.key];
 });
}

function mcpSettingsPanel(){
 const c=state.mcpConnection||{};
 const labels={ready:["Ready（控制平面检查通过）","Ready (control-plane check passed)"],client_missing:["未安装 Tunnel Client","Tunnel Client missing"],unconfigured:["未配置","Unconfigured"],stopped_or_unhealthy:["已配置，未通过健康检查","Configured; health check not passed"]};
 const button=(action,label)=>'<button type="button" class="btn" data-action="settings-mcp-'+action+'">'+E(settingsLabel(label))+'</button>';
 return '<p>'+E(settingsLabel(labels[c.state]||["请刷新连接状态","Refresh connection status"]))+'</p>'+
 '<div class="settings-field"><label>Tunnel ID<input id="mcp-tunnel-id" maxlength="80" value="'+E(c.tunnel_id||'')+'" placeholder="tunnel_…"></label></div>'+
 '<div class="btn-row">'+button("install-client",["安装官方客户端","Install official client"])+button("setup",["配置并启动","Configure and start"])+button("start",["启动","Start"])+button("stop",["停止","Stop"])+button("restart",["重启","Restart"])+button("refresh",["刷新状态","Refresh status"])+ '</div>'+
 '<p class="muted">'+E(settingsLabel(["首次凭据只在本机终端输入；已有凭据复用。Ready 表示 Tunnel 健康，不代表 ChatGPT 已完成授权。","First-time credentials stay in the local terminal; saved credentials are reused. Ready describes Tunnel health; ChatGPT authorization is separate."]))+'</p>'+
 '<h3>'+E(settingsLabel(["本地 MCP 客户端","Local MCP client"]))+'</h3><pre>'+E(JSON.stringify(c.local_config||{},null,2))+'</pre>'+button("copy",["复制连接配置","Copy configuration"]);
}
