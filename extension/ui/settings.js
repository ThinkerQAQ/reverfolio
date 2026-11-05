"use strict";

/*
 * Recursive standalone settings navigator.  Presentation pattern inspired by
 * familiar desktop setting panels, implemented independently for Reverfolio.
 * Leaf controls are bound only to Reverfolio's allowlisted settings API.
 */
const settingTree = [
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
  {id:"mcp-connection",name:["MCP 连接","MCP connection"],desc:["按引导连接 ChatGPT 或本地 AI 客户端","Connect ChatGPT or a local AI client"],kind:"mcp-connection"},
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
  '<div class="settings-savebar"><span class="settings-hint">'+E(settingsLabel(node.fields.some(f=>["server.listen","storage.database","github.auth"].includes(f.key))?["保存配置后需重启服务；改监听地址还需重新加载扩展。","Restart the service after saving; listener changes also require extension reload."]:["修改在保存后生效","Changes take effect after saving"]))+'</span>'+
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
  '<span class="settings-category-meta">'+(node.children?settingLeaves(node.children):'')+' <span class="settings-category-chevron">›</span></span></button>'
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

function mcpActive(){return !!state.data&&state.view==="settings"&&state.settingsSelected==="mcp-connection";}
function settingsMCPBeforeRender(){globalThis.ReverfolioMCPConnection?.capture();}
function settingsMCPAfterRender(){
 if(mcpActive())globalThis.ReverfolioMCPConnection?.init();
 else globalThis.ReverfolioMCPConnection?.unmount();
}
const mcpTranslations={"本地服务运行中 · Tunnel 未连接": "Local service running · Tunnel disconnected", "本地服务正在运行，但 Tunnel 未连接。点击「连接 Tunnel」，等待就绪后再到 ChatGPT 管理插件。": "The local service is running. Connect the Tunnel and wait until ready before managing the plugin in ChatGPT.", "无法读取系统自启动设置，请检查系统权限或登录会话。": "Unable to read system startup settings. Check system permissions or your login session.", "Tunnel 自动连接": "Connect Tunnel automatically", "本地服务启动后自动连接 Tunnel，修改后重启当前服务生效。": "Connect the Tunnel when the local service starts. Restart the service to apply changes.", "连接 Tunnel": "Connect Tunnel", "断开 Tunnel": "Disconnect Tunnel", "Tunnel 已断开，本地服务仍运行。": "Tunnel disconnected. The local service is still running.", "Tunnel 连接请求已提交，正在检查状态。": "Tunnel connection requested. Checking status.", "已保存。重启服务后，新的自动连接设置生效。": "Saved. Restart the service to apply automatic connection settings.", "Tunnel 自动连接设置已保存。": "Tunnel automatic connection setting saved.", "正在检查…": "Checking…", "本地 Tunnel 服务": "Local Tunnel service", "保存连接配置后启动服务；服务运行时，重启可应用已保存的修改。": "Save your connection settings, then start the service. Restart applies saved changes.", "尚未保存": "Not saved", "下次登录系统时启动。保存不会立即启动或停止当前服务。": "Start at the next sign-in. Saving does not start or stop the current service.", "ChatGPT 插件": "ChatGPT plugin", "本地 Tunnel 就绪后，在 ChatGPT 中创建并启用插件。授权结果以 ChatGPT 为准。": "Once the local Tunnel is ready, create and enable the plugin in ChatGPT. Authorization is handled there.", "本地 MCP 客户端无需 Tunnel ID 或 Runtime API Key。": "Local MCP clients do not need a Tunnel ID or Runtime API Key.", "配置保存在本机，已保存的 API Key 不会返回此页面。": "Settings are kept locally. Saved API keys are never returned to this page.", "创建或管理 Tunnel": "Create or manage a Tunnel", "创建 Runtime API Key": "Create a Runtime API Key", "显示输入的密钥": "Show entered key", "隐藏输入的密钥": "Hide entered key", "已保存的修改等待重启生效。重启服务后再到 ChatGPT 中管理插件。": "Saved changes require a restart. Restart the service before managing the plugin in ChatGPT.", "本地 Tunnel 已就绪。请在 ChatGPT 中创建或管理插件，授权结果以 ChatGPT 为准。": "The local Tunnel is ready. Create or manage the plugin in ChatGPT; authorization is handled there.", "先启动本地 Tunnel 并等待就绪，再到 ChatGPT 中创建或管理插件。": "Start the local Tunnel and wait until it is ready before managing the plugin in ChatGPT.", "本地服务开机自启": "Start local service at login", "启动服务": "Start Service", "重启服务": "Restart Service", "本地 MCP 客户端（STDIO）": "Local MCP clients (STDIO)", "复制本地配置": "Copy local configuration", "打开 ChatGPT 插件设置": "Open ChatGPT plugin settings", "连接": "Connect", "创建": "Create", "保存": "Save", "刷新": "Refresh", "复制": "Copy", "失败": "Failed", "更新": "Update", "检测": "Detect", "状态": "Status", "设置": "Settings", "未配置": "Not configured", "已更新": "Updated", "运行中": "Running", "正在保存": "Saving", "启用": "Enabled", "等待": "Waiting", "连接状态": "Connection status", "未连接": "Disconnected", "保存失败，之前的配置已保留。": "Save failed. The previous configuration was retained."};
Object.assign(mcpTranslations,{"保存超时；请核对 Tunnel ID 和 Runtime API Key。": "Save timed out. Check the Tunnel ID and Runtime API Key.", "停止服务": "Stop Service", "原有配置已确认，无需再次写入。": "The saved configuration is unchanged.", "已保存；可修改后再次保存": "Saved; edit and save to replace", "已保存；留空则复用": "Saved; leave blank to reuse", "已停止": "Stopped", "已提交配置校验；保存成功前不会启动服务。": "Validating configuration. Saving does not start the service.", "开机自启设置已更新，不会影响当前服务。": "Login startup updated. The current service is unchanged.", "操作已提交，正在刷新服务状态。": "Action submitted. Refreshing service status.", "本地配置已复制。": "Local configuration copied.", "检测到非手动填入的 Tunnel ID，已恢复原值。": "Automatic Tunnel ID input was rejected. The saved value was restored.", "正在启动 / 未就绪": "Starting / not ready", "等待重启": "Restart required", "请输入 Runtime API Key。": "Enter a Runtime API Key.", "请输入有效的非测试 Tunnel ID。": "Enter a valid, non-example Tunnel ID.", "配置已保存。": "Configuration saved."});
Object.assign(mcpTranslations,{
 "官方 Tunnel Client 已安装":"Official Tunnel Client installed",
 "官方 Tunnel Client 尚未安装（按需下载）":"Official Tunnel Client not installed (download on demand)",
 "官方 Tunnel Client 已下载并校验；不会自动启动服务。":"Official Tunnel Client downloaded and verified; service remains stopped.",
 "安装官方 Tunnel Client":"Install official Tunnel Client"
});
globalThis.ReverfolioMCPText=text=>state.lang==="en"?(mcpTranslations[text]||text):text;
const mcpPanelMarkup="<section id=\"mcpConnectionSettings\" class=\"mcp-setup\">\n          <div class=\"mcp-workflow\">\n            <div class=\"mcp-workflow-header\">\n              <span>连接状态</span>\n              <span id=\"mcpConnectionStatus\" class=\"mcp-status-chip\" role=\"status\" aria-live=\"polite\" data-kind=\"idle\">\n                <span class=\"mcp-status-dot\" aria-hidden=\"true\"></span>\n                <span id=\"mcpConnectionStatusText\">正在检查…</span>\n              </span>\n            </div>\n            <section class=\"mcp-flow-section\" aria-labelledby=\"mcpLocalTitle\">\n              <div class=\"mcp-section-heading\">\n                <span class=\"mcp-step\" aria-hidden=\"true\">01</span>\n                <div>\n                  <h3 id=\"mcpLocalTitle\">本地 Tunnel 服务</h3>\n                  <p>保存连接配置后启动服务；服务运行时，重启可应用已保存的修改。</p>\n                </div>\n              </div>\n              <div class=\"mcp-flow-body\">\n                <div class=\"mcp-field\">\n                  <div class=\"mcp-field-heading\">\n                    <label for=\"mcpTunnelID\">Tunnel ID</label>\n                    <a class=\"mcp-help\" href=\"https://platform.openai.com/settings/organization/tunnels\" target=\"_blank\" rel=\"noopener noreferrer\" title=\"创建或管理 Tunnel\" aria-label=\"创建或管理 Tunnel\">?</a>\n                  </div>\n                  <input id=\"mcpTunnelID\" name=\"reverfolio-tunnel-id\" type=\"text\" spellcheck=\"false\" autocomplete=\"off\" maxlength=\"80\" readonly>\n                  <small id=\"mcpTunnelHint\">尚未保存</small>\n                </div>\n                <div class=\"mcp-field\">\n                  <div class=\"mcp-field-heading\">\n                    <label for=\"mcpRuntimeKey\">Runtime API Key</label>\n                    <a class=\"mcp-help\" href=\"https://platform.openai.com/settings/organization/api-keys\" target=\"_blank\" rel=\"noopener noreferrer\" title=\"创建 Runtime API Key\" aria-label=\"创建 Runtime API Key\">?</a>\n                  </div>\n                  <div class=\"mcp-key-input\">\n                    <input id=\"mcpRuntimeKey\" name=\"reverfolio-tunnel-key\" type=\"password\" spellcheck=\"false\" autocomplete=\"new-password\" maxlength=\"512\" readonly>\n                    <button id=\"mcpKeyReveal\" class=\"mcp-reveal\" type=\"button\" disabled aria-pressed=\"false\" aria-label=\"显示输入的密钥\" title=\"显示输入的密钥\">\n                      <svg viewBox=\"0 0 24 24\" width=\"18\" height=\"18\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.7\" aria-hidden=\"true\">\n                        <path d=\"M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12Z\"/>\n                        <circle cx=\"12\" cy=\"12\" r=\"3\"/>\n                      </svg>\n                    </button>\n                  </div>\n                  <small id=\"mcpRuntimeHint\">尚未保存</small>\n                </div>\n                <label class=\"mcp-preference\" for=\"mcpAutoStart\">\n                  <input id=\"mcpAutoStart\" type=\"checkbox\" disabled>\n                  <span><strong>本地服务开机自启</strong><small>下次登录系统时启动。保存不会立即启动或停止当前服务。</small></span>\n                </label>\n                <small id=\"mcpStartupError\" class=\"mcp-startup-error\" role=\"status\" aria-live=\"polite\"></small>\n                <label class=\"mcp-preference mcp-auto-connect\" for=\"mcpAutoConnect\">\n                  <input id=\"mcpAutoConnect\" type=\"checkbox\" checked disabled>\n                  <span><strong>Tunnel 自动连接</strong><small>本地服务启动后自动连接 Tunnel，修改后重启当前服务生效。</small></span>\n                </label>\n                <div class=\"mcp-button-row\">\n                  <button id=\"mcpSave\" class=\"mcp-primary\" type=\"button\">保存</button>\n                  <button id=\"mcpConnect\" class=\"mcp-secondary\" type=\"button\" disabled>启动服务</button>\n                  <button id=\"mcpRestart\" class=\"mcp-secondary\" type=\"button\" disabled>重启服务</button>\n                </div>\n                <div class=\"mcp-client-actions\">\n                  <button id=\"mcpInstallClient\" class=\"mcp-secondary\" type=\"button\" hidden>安装官方 Tunnel Client</button>\n                  <small id=\"mcpClientHint\" role=\"status\">官方 Tunnel Client 尚未安装（按需下载）</small>\n                </div>\n                <div class=\"mcp-tunnel-actions\">\n                  <button id=\"mcpTunnelToggle\" class=\"mcp-secondary\" type=\"button\" disabled>连接 Tunnel</button>\n                </div>\n                <p id=\"mcpConnectionMessage\" class=\"mcp-message\" role=\"status\" aria-live=\"polite\"></p>\n              </div>\n            </section>\n            <section class=\"mcp-flow-section mcp-plugin-panel\" aria-labelledby=\"mcpPluginTitle\">\n              <div class=\"mcp-section-heading\">\n                <span class=\"mcp-step\" aria-hidden=\"true\">02</span>\n                <div>\n                  <h3 id=\"mcpPluginTitle\">ChatGPT 插件</h3>\n                  <p id=\"mcpPluginDescription\">本地 Tunnel 就绪后，在 ChatGPT 中创建并启用插件。授权结果以 ChatGPT 为准。</p>\n                </div>\n              </div>\n              <div class=\"mcp-flow-body\">\n                <a class=\"mcp-plugin-link\" href=\"https://chatgpt.com/plugins\" target=\"_blank\" rel=\"noopener noreferrer\">\n                  打开 ChatGPT 插件设置\n                  <svg viewBox=\"0 0 24 24\" width=\"15\" height=\"15\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"1.8\" stroke-linecap=\"round\" stroke-linejoin=\"round\" aria-hidden=\"true\"><path d=\"M14 5h5v5M19 5l-9 9M18 13v5H5V6h5\"/></svg>\n                </a>\n                <details class=\"mcp-setup-advanced\">\n                  <summary>本地 MCP 客户端（STDIO）</summary>\n                  <p>本地 MCP 客户端无需 Tunnel ID 或 Runtime API Key。</p>\n                  <pre id=\"mcpLocalConfig\" class=\"mcp-local-config\"></pre>\n                  <button id=\"mcpCopyConfig\" class=\"mcp-secondary\" type=\"button\">复制本地配置</button>\n                </details>\n              </div>\n            </section>\n          </div>\n          <p class=\"mcp-footer\">配置保存在本机，已保存的 API Key 不会返回此页面。</p>\n        </section>";
function mcpSettingsPanel(){
 let markup=mcpPanelMarkup;
 if(state.lang==="en")for(const [zh,en] of Object.entries(mcpTranslations).sort((a,b)=>b[0].length-a[0].length))markup=markup.split(zh).join(E(en));
 return markup;
}

globalThis.ReverfolioMCPAPI={
 errorMessage:error=>String(error?.message||error),
 async send(type,input={}){
  if(type==='reverfolio.mcp.connection.get')return {connection:await api('/api/mcp-connection')};
  const payload={action:input.action};
  if(input.action==='save')Object.assign(payload,{
   tunnel_id:input.tunnelId,runtime_api_key:input.runtimeApiKey,reuse_key:input.reuseKey,
   replace:input.replace,previous_tunnel_id:input.previousTunnelId
  });
  return api('/api/mcp-connection',{method:'POST',body:JSON.stringify(payload)});
 }
};
'use strict';
(function(root){
 const PATTERN=/^tunnel_(?:[a-z0-9]{32}|[a-z0-9]{4}_[a-z0-9]{32})$/;
 const EXAMPLE='tunnel_0123456789abcdef0123456789abcdef';
 const $=id=>document.getElementById(id);
 const t=text=>root.ReverfolioMCPText(text);
 const state={status:{},busy:false,pending:null,editedID:false,editedKey:false,autoStart:false,autoKnown:false,initialized:false,stopPoll:null};
 const validID=id=>PATTERN.test(id)&&id!==EXAMPLE;
 const running=s=>Boolean(s.running||s.ready);
 const controls=(s,{id,key,busy=false,pending=false}={})=>{
  const changed=Boolean((id||'')!==(s.tunnel_id||'')||key);
  const stored=Boolean(s.configured&&s.runtime_key_saved&&validID(s.tunnel_id));
  const active=running(s);
  return {
   canSave:!busy&&!pending,
   canStart:!active&&stored&&!changed&&!busy&&!pending,
   canStop:active&&!busy&&!pending,
   canRestart:active&&stored&&!changed&&!busy&&!pending,
   canAutoStart:stored&&!changed&&!busy&&!pending,
   changed,active,stored
  };
 };
 function message(text,error=false) {
  if(!$('mcpConnectionMessage'))return;
  $('mcpConnectionMessage').textContent=text||'';
  $('mcpConnectionMessage').className=error?'mcp-message error':'mcp-message';
 }
 function form() {
  const input=$('mcpTunnelID'), key=$('mcpRuntimeKey');
  return {id:input?.value?.trim()||'',key:key?.value?.trim()||'',busy:state.busy,pending:!!state.pending};
 }
 function show() {
  if(!$('mcpConnectionSettings'))return;
  const s=state.status,c=controls(s,form()),run=running(s);
  $('mcpSave').disabled=!c.canSave;
  if($('mcpInstallClient')) {
   $('mcpInstallClient').hidden=!!s.client_installed;
   $('mcpInstallClient').disabled=state.busy||!!state.pending;
  }
  if($('mcpClientHint'))$('mcpClientHint').textContent=s.client_installed?t('官方 Tunnel Client 已安装'):t('官方 Tunnel Client 尚未安装（按需下载）');
  $('mcpConnect').disabled=run?!c.canStop:!c.canStart;
  $('mcpConnect').textContent=run?t('停止服务'):t('启动服务');
  $('mcpRestart').disabled=!c.canRestart;
  const connectPref=$('mcpAutoConnect');
  if(connectPref){
   connectPref.disabled=!c.canAutoStart||!s.configured;
   connectPref.checked=s.tunnel_auto_connect!==false;
  }
  const toggle=$('mcpTunnelToggle');
  if(toggle) {
   const connecting=Boolean(s.tunnel_connecting||s.ready);
   toggle.textContent=connecting?t('断开 Tunnel'):t('连接 Tunnel');
   toggle.disabled=state.busy||!!state.pending||!run||(!connecting&&c.changed)||(!connecting&&s.restart_required);
  }
  if($('mcpAutoStart')) {
   $('mcpAutoStart').disabled=!c.canAutoStart||!state.autoKnown;
   $('mcpAutoStart').checked=state.autoStart;
  }
  let label=t('未配置'),kind='disabled';
  if(state.pending){label=t('正在保存');kind='checking';}
  else if(s.restart_required&&run){label=t('等待重启');kind='checking';}
  else if(s.ready){label='Ready';kind='ok';}
  else if(run&&!s.tunnel_connecting){label=t('本地服务运行中 · Tunnel 未连接');kind='disabled';}
  else if(run){label=t('正在启动 / 未就绪');kind='checking';}
  else if(s.configured){label=t('已停止');kind='disabled';}
  $('mcpConnectionStatusText').textContent=label;
  const statusChip=$('mcpConnectionStatus');
  const chipKind=kind==='ok'?'ready':kind==='checking'?'pending':'idle';
  if(statusChip.dataset) statusChip.dataset.kind=chipKind;
  else if(statusChip.setAttribute) statusChip.setAttribute('data-kind',chipKind);
  $('mcpPluginDescription').textContent=s.restart_required&&run
    ? t('已保存的修改等待重启生效。重启服务后再到 ChatGPT 中管理插件。')
    : s.ready
      ? t('本地 Tunnel 已就绪。请在 ChatGPT 中创建或管理插件，授权结果以 ChatGPT 为准。')
      : run&&!s.tunnel_connecting
        ? t('本地服务正在运行，但 Tunnel 未连接。点击「连接 Tunnel」，等待就绪后再到 ChatGPT 管理插件。')
        : t('先启动本地 Tunnel 并等待就绪，再到 ChatGPT 中创建或管理插件。');
  $('mcpTunnelHint').textContent=s.configured?t('已保存；可修改后再次保存'):t('尚未保存');
  $('mcpRuntimeHint').textContent=s.runtime_key_saved?t('已保存；留空则复用'):t('尚未保存');
  const reveal=$('mcpKeyReveal');
  if(reveal) {
   reveal.disabled=!$('mcpRuntimeKey').value;
   if(reveal.disabled) {
    $('mcpRuntimeKey').type='password';
    reveal.setAttribute?.('aria-pressed','false');
    reveal.setAttribute?.('aria-label',t('显示输入的密钥'));
   }
  }
 }
 function intendedInput(el,value) {
  if(!el)return;
  el.value=value;
  el.readOnly=true;
  el.autocomplete='off';
  el.setAttribute('data-lpignore','true');
  el.setAttribute('data-1p-ignore','true');
  el.setAttribute('data-bwignore','true');
  const unlock=()=>{el.readOnly=false;};
  for(const evt of ['keydown','pointerdown','paste'])el.addEventListener(evt,unlock);
  el.addEventListener('focus',()=>{
    const intended=el.id==='mcpTunnelID'
      ? (state.editedID?el.value:(state.status.tunnel_id||''))
      : (state.editedKey?el.value:'');
    if(el.value!==intended)el.value=intended;
  });
  el.addEventListener('input',()=>{
   if(document.activeElement!==el||el.matches?.(':-webkit-autofill')) {
    el.value=value;return;
   }
   if(el.id==='mcpTunnelID')state.editedID=true;
   else state.editedKey=true;
   message('');
   show();
  });
 }
 async function refresh(){
  if(state.busy)return;
  try{
   const result=await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.get');
   const s=result.connection||{};
   state.status=s;
   if(!$('mcpConnectionSettings'))return;
   if(!state.editedID&&document.activeElement!==$('mcpTunnelID'))$('mcpTunnelID').value=s.tunnel_id||'';
   $('mcpRuntimeKey').placeholder=s.runtime_key_saved?'****************':'';
   $('mcpLocalConfig').textContent=JSON.stringify(s.local_config||{},null,2);
   const pending=state.pending;
   if(pending){
    if(s.setup_state==='failed'){state.pending=null;message(t('保存失败，之前的配置已保留。'),true);}
    else if(s.setup_state==='succeeded'&&s.revision&&s.revision!==pending.revision&&s.tunnel_id===pending.id){
     state.pending=null;
     state.editedID=false;state.editedKey=false;
     $('mcpTunnelID').value=s.tunnel_id;
     $('mcpRuntimeKey').value='';
     message(t('配置已保存。'));
    }else if(Date.now()>pending.deadline){state.pending=null;message(t('保存超时；请核对 Tunnel ID 和 Runtime API Key。'),true);}
   }
   show();
  }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
 }
 async function action(action){
  if(state.busy||state.pending)return;
  if(action==='install-client'){
   state.busy=true;show();
   try{
    await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:'install-client'});
    message(t('官方 Tunnel Client 已下载并校验；不会自动启动服务。'));
   }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
   finally{state.busy=false;await refresh();}
   return;
  }
  const s=state.status,c=controls(s,form());
  if(action==='start'&&!c.canStart||action==='stop'&&!c.canStop||
     action==='restart'&&!c.canRestart)return;
  const id=$('mcpTunnelID').value.trim();
  // Ignore silent password-manager injection: only user-initiated input may rotate the key.
  const key=state.editedKey ? $('mcpRuntimeKey').value.trim() : '';
  if(action==='save'&&!state.editedID&&id!==(s.tunnel_id||'')){
   $('mcpTunnelID').value=s.tunnel_id||'';
   message(t('检测到非手动填入的 Tunnel ID，已恢复原值。'),true);
   show();return;
  }
  if(action==='save'){
   if(!validID(id)){message(t('请输入有效的非测试 Tunnel ID。'),true);return;}
   if(!key&&!s.runtime_key_saved){message(t('请输入 Runtime API Key。'),true);return;}
   if(id===s.tunnel_id&&!key&&s.configured&&s.runtime_key_saved){
    state.editedID=false;state.editedKey=false;$('mcpRuntimeKey').value='';
    message(t('原有配置已确认，无需再次写入。'));
    show();return;
   }
  }
  state.busy=true;show();
  try{
   if(action==='save'&&!s.client_installed){
    await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:'install-client'});
   }
   await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{
    action,tunnelId:id,runtimeApiKey:action==='save'?key:'',
    reuseKey:action==='save'&&!key,
    replace:action==='save'&&!!s.configured,previousTunnelId:s.tunnel_id||''
   });
   if(action==='save'){
    state.pending={id,revision:s.revision||'',deadline:Date.now()+120000};
    if($('mcpRuntimeKey'))$('mcpRuntimeKey').value='';
    message(t('已提交配置校验；保存成功前不会启动服务。'));
   }else message(t('操作已提交，正在刷新服务状态。'));
  }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
  finally{state.busy=false;await refresh();}
 }
 async function autoStatus(){
  try{
   const r=await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:'auto-status'});
   state.autoStart=!!r.auto_start;state.autoKnown=true;
    if($('mcpStartupError'))$('mcpStartupError').textContent='';
  }catch(e){
    state.autoKnown=false;
    if($('mcpStartupError'))$('mcpStartupError').textContent=t('无法读取系统自启动设置，请检查系统权限或登录会话。');
  }
  show();
 }
   async function autoConnectChanged(){
   if(state.busy||state.pending)return;
   const c=controls(state.status,form());
   if(!c.canAutoStart){show();return;}
   const wanted=!!$('mcpAutoConnect').checked;
   state.busy=true;show();
   try {
    await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:wanted?'auto-connect-enable':'auto-connect-disable'});
    message(state.status.running?t('已保存。重启服务后，新的自动连接设置生效。'):t('Tunnel 自动连接设置已保存。'));
   }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
   finally{state.busy=false;await refresh();}
  }
  async function tunnelChanged(){
   const s=state.status;
   if(state.busy||state.pending||!running(s))return;
   const connecting=Boolean(s.tunnel_connecting||s.ready);
   if(!connecting&&(!controls(s,form()).stored||controls(s,form()).changed||s.restart_required))return;
   state.busy=true;show();
   try {
    await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:connecting?'disconnect':'connect'});
    message(connecting?t('Tunnel 已断开，本地服务仍运行。'):t('Tunnel 连接请求已提交，正在检查状态。'));
   }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
   finally{state.busy=false;await refresh();}
  }
async function autoChanged(){
  if(state.busy||state.pending)return;
  const c=controls(state.status,form());
  if(!c.canAutoStart){show();return;}
  const desired=$('mcpAutoStart').checked;
  state.busy=true;show();
  try{
   const r=await root.ReverfolioMCPAPI.send('reverfolio.mcp.connection.action',{action:desired?'auto-enable':'auto-disable'});
   state.autoStart=!!r.auto_start;
   message(t('开机自启设置已更新，不会影响当前服务。'));
  }catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
  finally{state.busy=false;show();}
 }
 function init(){
  const panel=$('mcpConnectionSettings');
  if(state.panel===panel)return;
  state.stopPoll?.();state.panel=panel;state.initialized=true;
  const advanced=panel.querySelector?.('details');if(advanced)advanced.open=!!state.advanced;
  intendedInput($('mcpTunnelID'),state.editedID?state.draftID||'':state.status.tunnel_id||'');
  intendedInput($('mcpRuntimeKey'),state.editedKey?state.draftKey||'':'');
  $('mcpSave').addEventListener('click',()=>action('save'));
  $('mcpInstallClient')?.addEventListener('click',()=>action('install-client'));
  $('mcpConnect').addEventListener('click',()=>action(running(state.status)?'stop':'start'));
  $('mcpRestart').addEventListener('click',()=>action('restart'));
  $('mcpAutoStart')?.addEventListener('change',autoChanged);
  $('mcpAutoConnect')?.addEventListener('change',autoConnectChanged);
  $('mcpTunnelToggle')?.addEventListener('click',tunnelChanged);
  $('mcpKeyReveal')?.addEventListener('click',()=>{
   const input=$('mcpRuntimeKey'),reveal=$('mcpKeyReveal');
   if(!input.value || reveal.disabled)return;
   const showing=input.type==='password';
   input.type=showing?'text':'password';
   reveal.setAttribute('aria-pressed',String(showing));
   reveal.setAttribute('aria-label',showing?t('隐藏输入的密钥'):t('显示输入的密钥'));
   reveal.title=showing?t('隐藏输入的密钥'):t('显示输入的密钥');
  });
  $('mcpCopyConfig').addEventListener('click',async()=>{
   try{await navigator.clipboard.writeText(JSON.stringify(state.status.local_config||{},null,2));message(t('本地配置已复制。'));}
   catch(e){message(root.ReverfolioMCPAPI.errorMessage(e),true);}
  });
  const poll=root.setInterval(()=>{
   if(mcpActive())refresh();
  },4000);
  state.stopPoll=()=>root.clearInterval(poll);
  refresh();autoStatus();
 }
 root.ReverfolioMCPConnection={init,refresh,controls,validID,
  capture(){if(!state.panel?.isConnected)return;state.advanced=!!state.panel.querySelector?.('details')?.open;state.draftID=$('mcpTunnelID').value;state.draftKey=$('mcpRuntimeKey').value;},
  unmount(){state.stopPoll?.();state.panel=null;}
 };
})(globalThis);
