import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import vm from "node:vm";
const source=readFileSync(new URL("../internal/echo/web/dashboard.js",import.meta.url),"utf8");
const listeners={};
let saved;
const context={settingsLabel:p=>p[1],state:{data:{works:[]}},E:x=>String(x??"").replaceAll("<","&lt;"),fmt:x=>String(x),dt:x=>String(x),heading:x=>"<h2>"+x+"</h2>",document:{addEventListener:(event,fn)=>listeners[event]=fn,getElementById:()=>({})},blobSave:(blob,name)=>saved={blob,name},Blob,Date,Number,Map,Set,render:()=>{},notification:err=>{throw Error(err)}};
vm.createContext(context);vm.runInContext(source,context);
const html=context.metricChart("Synthetic",[{label:"2026-10-01",value:0},{label:"2026-10-03",value:5}],"UTC");
assert.match(html,/2026-10-01: 0/);
assert.match(html,/2026-10-03: 5/);
assert.ok(!html.includes("2026-10-02"),"Never invent missing buckets");
assert.match(html,/role="img"/);
assert.match(context.metricChart("Missing",[]),/missing is not zero/);
const button={dataset:{chartExport:"impact-chart-1",format:"csv"}};
await listeners.click({target:{closest:()=>button}});
assert.equal(saved.name,"impact-chart-1.csv");
assert.match(await saved.blob.text(),/"2026-10-01","0"/);
context.state.data.dashboard={projects:[],providers:[{source_id:"fixture",provider:"gsc",state:"failed",captured_at:"2026-10-31",problem:"provider_http_503",metrics:[],details:[]}]};
const dashboard=context.impactDashboardView();
assert.match(dashboard,/provider_http_503/);
assert.match(dashboard,/exact normalized URLs/);
assert.match(html,/viewBox="0 0 760 266"/,"SVG export must include the descriptive footer");
assert.match(html,/<desc>UTC<\/desc>/);
context.state.chartRange="7";
context.state.data.dashboard.projects=[{
 work:{id:1,title:"Demo",repository:"demo/one"},
 attempt_at:"2026-10-20T00:00:00Z",attempt_state:"complete",
 traffic:{},days:{views:[],clones:[]},popular:{},
 observations:[
  {at:"2026-10-20T00:00:00Z",stars:4},
  {at:"2026-10-01T00:00:00Z",stars:3}
 ],
 changes:{from:"2026-10-01T00:00:00Z",to:"2026-10-20T00:00:00Z",known_downloads:{total:2,packages:1},unknown_assets:1}
}];
const chartHTML=context.impactDashboardView();
assert.match(chartHTML,/Observed download change/);
assert.match(chartHTML,/unknown change/);
const chartValues=vm.runInContext("Array.from(chartRows.values())",context);
assert.ok(chartValues.some(rows=>rows.length===1&&rows[0].value===4),"Seven days must mean calendar dates, not the last seven snapshots");
assert.ok(!chartValues.some(rows=>rows.some(r=>r.value===3)),"Do not leak out-of-range cumulative samples");
context.state.data.works=[{id:12,title:"Article",kind:"article"}];
context.state.chartWork="12";
context.state.data.dashboard.providers=[
 {source_id:"matching",work_id:11,provider:"umami",state:"complete",captured_at:"2026-10-20",metrics:[],details:[{work_id:12,metric:"pageviews",value:5,page_url:"https://example.org/article",period_start:"2026-10-01",period_end:"2026-10-20"}]},
 {source_id:"matching",work_id:11,provider:"umami",state:"failed",captured_at:"2026-10-21",problem:"provider_timeout",metrics:[],details:[]},
 {source_id:"unrelated",work_id:11,provider:"umami",state:"complete",captured_at:"2026-10-20",metrics:[],details:[{work_id:13,metric:"pageviews",value:6,page_url:"https://example.org/other",period_start:"2026-10-01",period_end:"2026-10-20"}]}
];
const articleHTML=context.providerCharts(context.state.data.dashboard.providers);
assert.match(articleHTML,/matching/);
assert.match(articleHTML,/provider_timeout/,"Keep latest failed attempt alongside successful attributed data");
assert.ok(!articleHTML.includes("unrelated"),"Article filter must not display unrelated provider data");
context.state.chartWork="13";
assert.match(context.providerCharts([]),/No site observations/);
context.state.chartProject="";
const noData={work:{id:1,title:"Fail",repository:"demo/bad"},attempt_state:"failed",traffic:{},days:{views:[],clones:[]},popular:{},observations:[]};
const healthy={...noData,work:{id:2,title:"Healthy",repository:"demo/healthy"},attempt_state:"complete",traffic_at:"2026-10-21T00:00:00Z"};
context.state.data.dashboard.projects=[noData,healthy];
assert.match(context.impactDashboardView(),/<strong>demo\/healthy<\/strong>/,"First load should prefer a project with successful observations");
console.log("Charts preserve zero, missing days, true calendar ranges, article scope, interval deltas, explanatory SVG and downloadable CSV.");
