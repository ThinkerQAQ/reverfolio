import assert from "node:assert/strict";
import fs from "node:fs";
import vm from "node:vm";
const source=fs.readFileSync(new URL("../internal/echo/web/settings.js",import.meta.url),"utf8");
const calls=[],timers=new Map(),elements=new Map();
let id=0,fail=false;
const connection={state:"client_missing",client_installed:false,configured:false,ready:false,tunnel_id:"",local_config:{mcpServers:{reverfolio:{command:"fixture"}}}};
const context={
 state:{lang:"en",view:"settings",settingsSelected:"mcp-connection",data:{},mcpConnection:{...connection}},
 E:value=>String(value??"").replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll('"',"&quot;"),
 document:{getElementById:key=>elements.get(key)},
 navigator:{clipboard:{writeText:async()=>{}}},notification:()=>{},
 setTimeout:fn=>{timers.set(++id,fn);return id;},clearTimeout:key=>timers.delete(key),
 render:()=>{},
 api:async(url,options={})=>{
  calls.push([url,options]);
  if(fail)throw Error("network unavailable");
  if(options.method==="POST"){
   const action=JSON.parse(options.body).action;
   if(action==="install-client")connection.client_installed=true;
   if(action==="setup")connection.configured=true;
   if(action==="stop")connection.ready=false;
   return {};
  }
  return {...connection};
 }
};
vm.createContext(context);vm.runInContext(source,context);
const ui=vm.runInContext("mcpUI",context);
const main=()=>context.mcpSettingsPanel().split('<details')[0];
assert.equal((main().match(/<button/g)||[]).length,1);
assert.match(main(),/Connect ChatGPT/);
assert.ok(!main().includes("JSON"));
assert.ok(!context.mcpSettingsPanel().includes('id="mcp-advanced" class="mcp-advanced" open'));
ui.draftID="invalid";
await context.mcpHandleAction("connect");
assert.equal(calls.length,0,"Reject invalid IDs before downloading or changing anything");
assert.match(ui.error,/complete Tunnel ID/);
ui.draftID="tunnel_"+"0".repeat(32);
await context.mcpHandleAction("connect");
assert.deepEqual(calls.filter(([,o])=>o.method==="POST").map(([,o])=>JSON.parse(o.body).action),["install-client","setup"]);
assert.equal(ui.busy,"waiting");
assert.match(main(),/disabled/);
assert.equal(ui.draftID,"tunnel_"+"0".repeat(32));
connection.ready=true;
await context.mcpReadStatus();
assert.equal(ui.busy,"");
assert.match(main(),/Your local connection is ready/);
assert.match(main(),/complete authorization/,"Local readiness must not claim ChatGPT authorization");
assert.equal((main().match(/<button/g)||[]).length,0);
connection.ready=false;context.state.mcpConnection={...connection};
calls.length=0;
await context.mcpHandleAction("connect");
assert.deepEqual(calls.filter(([,o])=>o.method==="POST").map(([,o])=>JSON.parse(o.body).action),["start"]);
ui.deadline=0;
await context.mcpReadStatus();
assert.equal(ui.busy,"");
assert.match(ui.error,/not ready/);
ui.error="";fail=true;
await context.mcpHandleAction("connect");
assert.equal(ui.busy,"");
assert.match(ui.error,/network unavailable/);
assert.equal(ui.draftID,"tunnel_"+"0".repeat(32),"Retry must preserve input");
fail=false;
elements.set("mcp-tunnel-id",{value:"typed draft"});
elements.set("mcp-advanced",{open:true});
context.settingsMCPBeforeRender();
assert.equal(ui.draftID,"typed draft");
assert.equal(ui.advanced,true);
context.settingsMCPAfterRender();
assert.ok(ui.timer);
context.state.view="overview";
calls.length=0;
await context.mcpPoll();
assert.equal(calls.length,0,"Do not poll after leaving the MCP page");
context.settingsMCPAfterRender();
assert.equal(ui.timer,null);
console.log("MCP guide: one next action, automatic dependencies/readiness, preserved input, retries and scoped polling.");
