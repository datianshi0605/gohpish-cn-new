const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
for (const file of ['static/js/src/app/campaign_results.js', 'static/js/dist/app/campaign_results.min.js']) {
    const rows = [['old-rid', 'caret', 'Old', '', 'old@example.com', '', 'Email Sent', false, 'original-date']];
    let draws = 0, expandedUpdates = 0, visits = 0;
    const table = {
        rows() { return {every(fn) { rows.forEach((_, i) => fn.call(table, i)); }}; },
        row(i) {
            const child = function () { expandedUpdates++; };
            child.isShown = () => i === 0;
            return {data(value) { if (value) rows[i] = value; return rows[i]; }, child, node() { return {}; }};
        },
        draw(reset) { assert.equal(reset, false); draws++; }
    };
    table.row.add = row => rows.push(row);
    const chain = {ready() {}, hide() {}, show() {}, tooltip() {}, find() {return this;}, removeClass() {return this;}, addClass() {return this;},
        DataTable: () => table, highcharts: () => ({series: [{update() {}}]})};
    const $ = () => chain;
    $.each = (items, fn) => { for (const key of Object.keys(items)) { visits++; if (fn(key, items[key]) === false) break; } };
    let payload = {id: 1, timeline: [], results: [
        {id: 'old-rid', email: 'old@example.com', status: 'Email Opened', reported: true, send_date: 'scheduled'},
        {id: 'new-rid', email: 'new@example.com', first_name: '<New>', status: 'Scheduled', reported: false, send_date: 'later'}
    ]};
    const moment = value => ({format: () => value});
    const context = { $, document: {}, moment, escapeHtml: s => (s || '').replace(/</g, '&lt;').replace(/>/g, '&gt;'),
        api: {campaignId: {results: () => ({success: fn => fn(payload)})}}};
    vm.createContext(context);
    vm.runInContext(fs.readFileSync(path.join(__dirname, '..', file), 'utf8'), context);
    context.campaign = {id: 1}; context.updateMap = () => {}; context.renderTimeline = () => 'timeline';
    context.poll();
    assert.equal(rows.length, 2, 'poll must display newly synchronized recipients');
    assert.equal(rows[0][6], 'Email Opened');
    assert.equal(rows[0][7], true);
    assert.equal(rows[1][0], 'new-rid');
    assert.equal(rows[1][2], '&lt;New&gt;');
    assert.equal(rows[1][8], 'later');
    context.poll();
    assert.equal(rows.length, 2, 'repeat polls must not duplicate rows');
    assert.equal(expandedUpdates, 2, 'keep expanded history updated');
    assert.equal(draws, 2);
    payload.results.push({id: 'third-rid', email: 'third@example.com', status: 'Scheduled', send_date: 'later'});
    context.poll();
    assert.equal(rows.length, 3);
    for (let i=0;i<1200;i++) payload.results.push({id:'bulk-'+i,email:'bulk'+i+'@example.com',status:'Scheduled',send_date:'later'});
    context.poll();
    assert.equal(rows.length,1203);
    visits=0;context.poll();
    assert.equal(rows.length,1203);
    assert.ok(visits < 4 * 1203 + 20, 'refresh must not scan all results for every displayed row');
    console.log(file + ': recipient refresh, deduplication, history, escaping, 1203-row linear lookup PASS');
}
