const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const escapeHtml = text => text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
let shown = false, content = [];
const summary = {hide() {shown=false;return this;}, show(){shown=true;return this;},empty(){content=[];return this;},append(value){content.push(typeof value==='string'?value:value.value);return this;}};
function $(selector){
 if(selector==='#campaignModeSummary')return summary;
 if(selector==='<span>')return {value:'',addClass(){return this;},text(value){this.value=value;return this;}};
 return {ready(){}};
}
const context = {$, escapeHtml, document:{createTextNode:value=>value}};
vm.createContext(context);
vm.runInContext(fs.readFileSync(path.join(__dirname,'../static/js/src/app/campaigns.js'),'utf8'),context);
assert.match(context.campaignModeName({name:'普通活动',long_term:false}),/普通演练/);
assert.match(context.campaignModeName({name:'<script>',long_term:true,status:'In progress'}),/&lt;script&gt;.*长期演练/);
assert.match(context.campaignModeName({name:'结束',long_term:true,status:'Completed'}),/长期演练 · 已结束/);
vm.runInContext(fs.readFileSync(path.join(__dirname,'../static/js/src/app/campaign_results.js'),'utf8'),context);
context.campaignLongTermMode = true;
context.renderCampaignMode({status:'In progress'});
assert.equal(shown,true);assert.match(content.join(''),/每分钟检查新人/);
context.renderCampaignMode({status:'Completed'});
assert.match(content.join(''),/已结束/);assert.doesNotMatch(content.join(''),/每分钟/);
context.campaignLongTermMode=false;context.renderCampaignMode({status:'In progress'});assert.equal(shown,true);assert.match(content.join(''),/普通演练/);assert.doesNotMatch(content.join(''),/每分钟/);
console.log('Active/completed/ordinary campaign badges and escaped names: PASS');
