"use strict";
// Native SVG charts: exact observed values, no CDN, interpolation or invented zeros.
const D=(zh,en)=>settingsLabel([zh,en]);
const chartRows=new Map();
let chartSequence=0;
function metricChart(title, rows, note="") {
 const id="impact-chart-"+(++chartSequence);
 rows=rows.filter(r=>Number.isFinite(r.value));
 chartRows.set(id,rows);
 if(!rows.length)return '<section class="panel chart-panel"><h3>'+E(title)+'</h3><p class="muted">'+E(D("尚无观测数据；缺失不代表零。","No observations; missing is not zero."))+'</p></section>';
 const max=Math.max(1,...rows.map(r=>r.value)),width=760,height=250,left=64,plot=660,step=plot/rows.length;
 const noteLines=[];let line="",units=0;for(const ch of String(note)){const size=ch.charCodeAt(0)>255?2:1;if(units+size>105){noteLines.push(line);line="";units=0;}line+=ch;units+=size;}if(line)noteLines.push(line);
 const chartHeight=250+Math.min(4,noteLines.length)*16;
 const svg='<svg id="'+id+'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 '+chartHeight+'" role="img" aria-labelledby="'+id+'-title"><title id="'+id+'-title">'+E(title)+'</title><desc>'+E(note)+'</desc><rect width="760" height="'+chartHeight+'" fill="#ffffff"/><text x="64" y="24" font-size="16" fill="#152447">'+E(title)+'</text>'+
 [0,.5,1].map(f=>'<line x1="64" y1="'+(205-f*160)+'" x2="735" y2="'+(205-f*160)+'" stroke="#dce3ed"/><text x="56" y="'+(209-f*160)+'" text-anchor="end" font-size="12" fill="#526078">'+E(fmt(max*f))+'</text>').join('')+
 rows.map((r,i)=>{const x=left+i*step+step*.1,h=r.value/max*160;return '<rect x="'+x+'" y="'+(205-h)+'" width="'+Math.max(.8,step*.8)+'" height="'+h+'" fill="#4f61d6"><title>'+E(r.label+': '+r.value)+'</title></rect>'+
 (i===0||i===rows.length-1||rows.length<=8?'<text x="'+(x+step*.4)+'" y="228" text-anchor="middle" font-size="11" fill="#526078">'+E(r.label.length>22?r.label.slice(0,20)+'…':r.label)+'</text>':'');}).join('')+noteLines.slice(0,4).map((text,i)=>'<text x="64" y="'+(252+i*16)+'" font-size="10" fill="#526078">'+E(text)+(i===3&&noteLines.length>4?'…':'')+'</text>').join('')+'</svg>';
 return '<section class="panel chart-panel"><div class="work-header"><h3>'+E(title)+'</h3><div class="btn-row"><button class="btn btn-sm" data-chart-export="'+id+'" data-format="svg">SVG</button><button class="btn btn-sm" data-chart-export="'+id+'" data-format="png">PNG</button><button class="btn btn-sm" data-chart-export="'+id+'" data-format="csv">CSV</button></div></div>'+svg+'<p class="meta">'+E(note)+'</p><details><summary>'+E(D("查看精确数值","Exact values"))+'</summary><div class="table-wrap"><table class="data-table"><thead><tr><th>'+E(D("日期／分类","Date / category"))+'</th><th>'+E(D("数值","Value"))+'</th></tr></thead><tbody>'+rows.map(r=>'<tr><td>'+E(r.label)+'</td><td>'+E(fmt(r.value))+'</td></tr>').join('')+'</tbody></table></div></details></section>';
}
function impactDashboardView(){
 chartRows.clear();chartSequence=0;
 const data=state.data.dashboard;
 if(!data)return '';
 const projects=data.projects||[];
 const id=String(state.chartProject||projects[0]?.work.id||"");
 const project=projects.find(p=>String(p.work.id)===id)||projects[0];
 const range=state.chartRange||"14";
 const toolbar='<div class="toolbar"><label>'+E(D("项目","Project"))+' <select class="inline-input" data-chart-select="project">'+projects.map(p=>'<option value="'+p.work.id+'" '+(p===project?'selected':'')+'>'+E(p.work.title)+'</option>').join('')+'</select></label><label>'+E(D("最近已观测日期起","From last observed day"))+' <select class="inline-input" data-chart-select="range">'+[["7","7"],["14","14"],["all",D("全部历史","All history")]].map(([v,l])=>'<option value="'+v+'" '+(v===range?'selected':'')+'>'+E(l)+'</option>').join('')+'</select></label></div>';
 let html=heading(D("影响力统计图","Impact charts"))+toolbar;
 if(project){
  const p=project;
  html+='<div class="panel"><strong>'+E(p.work.repository)+'</strong><p>'+E(D("最近采集尝试","Latest attempt"))+': '+E(p.attempt_state||D("尚未采集","Not collected"))+' · '+dt(p.attempt_at)+'</p>'+((p.problems||[]).length?'<p class="warning">'+E(p.problems.join(' · '))+'</p>':'')+'<p class="meta">'+E(D("流量最近成功观测","Latest successful traffic observation"))+': '+dt(p.traffic_at)+' · '+E(D("下载观测","Download observation"))+': '+dt(p.download_at)+'</p></div><div class="impact-charts">';
  for(const kind of ["views","clones"]){
   const all=(p.days[kind]||[]).slice().sort((a,b)=>a.day.localeCompare(b.day));
   const last=all.at(-1)?.day;
   const cutoff=last&&range!=="all"?new Date(Date.parse(last.slice(0,10)+"T00:00:00Z")-(Number(range)-1)*86400000).toISOString().slice(0,10):"";
   const days=all.filter(d=>!cutoff||d.day.slice(0,10)>=cutoff);
   const label=kind==="views"?D("仓库页面浏览","Repository views"):D("仓库克隆","Repository clones");
   const traffic=p.traffic[kind];
   const note=D("GitHub · UTC 日数据；空缺不补零。","GitHub · UTC days; missing days remain absent.")+(traffic?' '+D("最近一次 API 窗口总数／独立标识","Latest API window count / unique identifiers")+': '+traffic.count+' / '+traffic.uniques:'');
   html+=metricChart(label,days.map(d=>({label:d.day.slice(0,10),value:d.count})),note);
   html+=metricChart(label+' · '+D("每日独立标识","Daily unique identifiers"),days.map(d=>({label:d.day.slice(0,10),value:d.uniques})),D("每日独立标识不可相加为跨日独立人数；克隆不等于使用。","Daily uniques cannot be added into cross-day people; clones are not usage."));
  }
  for(const [key,title] of [["stars","Stars"],["forks","Forks"],["downloads",D("发布附件累计下载","Cumulative release asset downloads")]]){
   const rows=(p.observations||[]).filter(o=>o[key]!=null).slice().reverse().map(o=>({label:new Date(o.at).toISOString(),value:o[key]}));
   const last=rows.at(-1)?.label.slice(0,10)||"";
   const cutoff=last&&range!=="all"?new Date(Date.parse(last+"T00:00:00Z")-(Number(range)-1)*86400000).toISOString().slice(0,10):"";
   html+=metricChart(title,cutoff?rows.filter(row=>row.label.slice(0,10)>=cutoff):rows,D("按采集时间的累计观测；范围为最后一次观测起的日历天数，不补缺失。下载不代表安装人数。","Cumulative observations within calendar days counted back from the latest observation, with no interpolation. Downloads are not installations."));
  }
  if(p.changes?.known_downloads&&p.changes.from&&p.changes.to){
   const c=p.changes;
   html+='<section class="panel"><h3>'+E(D("实际观测区间的下载增量","Observed download change"))+'</h3><p><strong>+'+fmt(c.known_downloads.total)+'</strong> '+E(D("已知附件下载增量","known asset downloads"))+' · '+E(D("安装包","Packages"))+': +'+fmt(c.known_downloads.packages)+'</p><p class="meta">'+dt(c.from)+' → '+dt(c.to)+' · '+E(D("无法确认增量的附件","Assets with unknown change"))+': '+fmt(c.unknown_assets)+'</p></section>';
  }
  html+=metricChart(D("按版本下载","Downloads by release"),p.releases||[],D("累计所有附件；旧快照缺版本时明确标记未知。","All cumulative assets; legacy snapshots without tags are labelled unknown."));
  html+=metricChart(D("按平台下载","Downloads by platform"),p.platforms||[],D("按附件名称推断；辅助文件和源码单列。","Inferred from filenames; auxiliary files and source are separate."));
  for(const kind of ["paths","referrers"]){
   const rows=p.popular[kind];
   html+=metricChart(kind==="paths"?D("热门仓库路径","Popular repository paths"):D("仓库流量来源","Repository referrers"),Array.isArray(rows)?rows.map(r=>({label:r.path||r.referrer||"?",value:r.count})):[],D("GitHub 最近成功窗口；不是博客页面访问。","GitHub latest available window; not blog page traffic."));
  }
  html+='</div>';
 }
 html+=providerCharts(data.providers||[]);
 const guide='<details class="panel"><summary>'+E(D("这些数据怎么用、怎么配置？","How to use and configure this data"))+'</summary><ol><li>'+E(D("在作品页登记你自己的仓库、网站或文章；只采集明确登记的对象。","Register your own repository, site or article under Works; only registered objects are monitored."))+'</li><li>'+E(D("给文章添加各平台已发布网址；按规范化后的完整 URL 精确归属，不按标题猜测。","Add each published platform URL. Attribution uses exact normalized URLs, never title guesses."))+'</li><li>'+E(D("在设置中启用 GitHub、GSC、Bing 或 Umami 数据源，选择所属作品、时区及日期窗口。凭据只保存在本地。","Enable GitHub, GSC, Bing or Umami under Settings, selecting work, timezone and dates. Credentials stay local."))+'</li><li>'+E(D("运行采集，先查看状态与日期，再比较同一来源、同一口径的趋势。CSV 用于分析，PNG/SVG 用于报告。","Collect, inspect status and dates, then compare trends within the same source and scope. Use CSV for analysis and PNG/SVG for reports."))+'</li></ol><p>'+E(D("文章清单来自你登记或导入的目录，来源在作品页可追溯。来源未知的历史记录不会自动当作已证实来源。","Articles come from registration or an imported catalog; provenance is shown under Works. Unknown legacy origins stay unknown."))+'</p></details>';
 return guide+html;
}
function providerCharts(runs){
 let html=heading(D("网站与文章统计","Site and article analytics"));
 const works=state.data.works.filter(w=>w.kind!=="project");
 const selected=String(state.chartWork||"");
 const selectedWork=works.find(w=>String(w.id)===selected);
 const wholeSite=!selected||selectedWork?.kind==="site";
 const articleSelected=selectedWork?.kind==="article";
 if(selectedWork?.kind==="site")runs=runs.filter(r=>String(r.work_id)===selected);
 if(articleSelected){
  const relevantSources=new Set(runs.filter(r=>(r.details||[]).some(m=>String(m.work_id)===selected)).map(r=>r.source_id));
  runs=runs.filter(r=>relevantSources.has(r.source_id));
 }
 html+='<div class="toolbar"><label>'+E(D("统计对象","Work"))+' <select class="inline-input" data-chart-select="work"><option value="">'+E(D("全部网站","All sites"))+'</option>'+works.map(w=>'<option value="'+w.id+'" '+(String(w.id)===selected?'selected':'')+'>'+E(w.title)+'</option>').join('')+'</select></label></div>';
 if(!runs.length)return html+'<p class="panel">'+E(articleSelected?D("此文章尚无可归属的页面统计。请先登记准确的原文及分发 URL，在设置中启用页面明细后重新采集。","No attributable page metrics for this article. Register the exact original and syndication URLs, enable page details in Settings and collect again."):D("尚无网站数据，请先在设置中配置数据源。","No site observations. Configure a source in Settings."))+'</p>';
 for(const r of runs){
  html+='<section class="panel"><h3>'+E(r.source_id)+' · '+E(r.state)+'</h3><p class="meta">'+dt(r.captured_at)+' · '+E(r.provider)+'</p>'+(r.problem?'<p>'+E(r.problem)+'</p>':'')+(r.detail_problem?'<p>'+E(D("明细采集问题","Detail collection problem"))+': '+E(r.detail_problem)+'</p>':'');
  const details=(r.details||[]).filter(m=>wholeSite?!m.page_url:String(m.work_id)===selected);
  const daily=details.length?details.filter(m=>Date.parse(m.period_end)-Date.parse(m.period_start)<=25*3600000):(!wholeSite?[]:r.metrics.filter(m=>Date.parse(m.period_end)-Date.parse(m.period_start)<=25*3600000));
  const keys=[...new Set(daily.map(m=>m.metric))];
  html+='<div class="impact-charts">';
  for(const key of keys){
   const dayMap=new Map();
   for(const m of daily.filter(m=>m.metric===key)) {const label=m.period_start.slice(0,10);dayMap.set(label,(dayMap.get(label)||0)+m.value);}
   html+=metricChart(key,[...dayMap].sort((a,b)=>a[0].localeCompare(b[0])).map(([label,value])=>({label,value})),daily.find(m=>m.metric===key)?.scope+" · "+daily.find(m=>m.metric===key)?.timezone);
  }
  if(!wholeSite){const windows=details.filter(m=>Date.parse(m.period_end)-Date.parse(m.period_start)>25*3600000);for(const m of windows)html+=metricChart(m.metric+' · '+D("页面窗口汇总","Page window total"),[{label:m.page_url,value:m.value}],m.period_start+' → '+m.period_end+' · '+m.scope);}
  if(wholeSite)for(const metric of ["search_clicks","search_impressions","pageviews"]){
   const pages=new Map();for(const m of (r.details||[]).filter(m=>m.page_url&&m.metric===metric))pages.set(m.page_url,(pages.get(m.page_url)||0)+m.value);
   if(pages.size)html+=metricChart(metric+' · '+D("页面排名","Page ranking"),[...pages].sort((a,b)=>b[1]-a[1]).slice(0,20).map(([label,value])=>({label,value})),D("仅提供方返回的页面；未返回不代表零。","Returned pages only; absent does not mean zero."));
  }
  html+='</div>';
  if(!daily.length)html+='<p class="muted">'+E(D("暂无可绘制的日明细；窗口汇总仍在下方来源卡中。文章明细需启用明细采集并精确登记页面 URL。","No daily detail available; window summaries remain in the source cards below. Enable details and register exact URLs for article attribution."))+'</p>';
  html+='</section>';
 }
 return html;
}
document.addEventListener("change",e=>{
 const key=e.target.dataset.chartSelect;if(!key)return;
 state[{project:"chartProject",range:"chartRange",work:"chartWork"}[key]]=e.target.value;render();
});
document.addEventListener("click",async e=>{
 const b=e.target.closest("[data-chart-export]");if(!b)return;
 try{
 const id=b.dataset.chartExport,format=b.dataset.format,svg=document.getElementById(id);
 if(format==="csv"){
  const quote=x=>'"'+String(x).replaceAll('"','""')+'"';
  const csv="label,value\r\n"+chartRows.get(id).map(r=>[r.label,r.value].map(quote).join(',')).join('\r\n');
  blobSave(new Blob([csv],{type:"text/csv;charset=utf-8"}),id+".csv");return;
 }
 const data=new XMLSerializer().serializeToString(svg);
 if(format==="svg"){blobSave(new Blob([data],{type:"image/svg+xml"}),id+".svg");return;}
 const url=URL.createObjectURL(new Blob([data],{type:"image/svg+xml"}));
 try{
  const img=new Image();img.src=url;await img.decode();
  const canvas=document.createElement("canvas");canvas.width=1520;canvas.height=Number(svg.viewBox.baseVal.height)*2;
  canvas.getContext("2d").drawImage(img,0,0,canvas.width,canvas.height);
  const blob=await new Promise(resolve=>canvas.toBlob(resolve,"image/png"));
  if(!blob)throw new Error("PNG export failed");
  const a=document.createElement("a"),downloadURL=URL.createObjectURL(blob);a.href=downloadURL;a.download=id+".png";a.click();setTimeout(()=>URL.revokeObjectURL(downloadURL),1000);
 } finally {URL.revokeObjectURL(url);}
 }catch(err){notification(err.message,true);}
});
