/* Long-term campaign controls; existing reporting and delivery UI is unchanged. */
$(function () {
    var id = window.location.pathname.split('/').pop();
    if (!id) { $('#longTermPanel').hide(); return; }
    function request(path, data) {
        return $.ajax({url:'/api/campaigns/' + id + path, method:data === undefined ? 'GET' : 'POST',
            contentType:'application/json', dataType:'json', data:data === undefined ? undefined : JSON.stringify(data),
            beforeSend:function (xhr) { xhr.setRequestHeader('Authorization', 'Bearer ' + user.api_key); }});
    }
    function failure(xhr) { $('#longTermFeedback').text(xhr.responseJSON && xhr.responseJSON.message || '操作失败，请稍后重试'); }
    request('').done(function (c) {
        $('#longTermEnabled').prop('checked', !!c.long_term);
        $('#autoGroupSettings').toggle(!!c.long_term && c.status !== 'Completed');
        if(c.status === 'Completed') { $('#longTermPanel input, #longTermPanel button').prop('disabled',true); $('#longTermFeedback').text('活动已完成，已停止自动加入新人员。'); }
    }).fail(failure);
    api.groups.get().success(function (groups) {
        groups.forEach(function(g) {
            $('<option>').val(g.id).text(g.name).appendTo('#autoGroups');
        });
        request('/auto-groups').done(function(selected) {
            $('#autoGroups').val(selected.map(function(g) { return String(g.id); }));
            if ($.fn.select2) { $('#autoGroups').select2({placeholder:'选择一个或多个用户组',width:'100%'}); }
            $('#saveAutoGroups').prop('disabled',false);
        }).fail(failure);
    });
    $('#saveLongTerm').on('click',function () {
        var enabled = $('#longTermEnabled').prop('checked');
        var button = $(this).prop('disabled',true);
        request('/long-term',{enabled:enabled}).done(function () {
            $('#autoGroupSettings').toggle(enabled);
            $('#longTermFeedback').text(enabled ? '长期演练已开启。请确认已选择并保存自动跟随的用户组，此后只需维护组成员。' : '已关闭自动同步，已有记录和计划投递保持不变。');
        }).fail(failure).always(function () { button.prop('disabled',false); });
    });
    $('#saveAutoGroups').on('click',function () {
        var ids=($('#autoGroups').val() || []).map(Number);
        var button=$(this).prop('disabled',true);
        request('/auto-groups',{group_ids:ids}).done(function () {
            $('#autoGroupsFeedback').text(ids.length ? '已关联 '+ids.length+' 个组。以后只需在这些组中添加人员，系统会自动加入本活动；已有人员不重复投递。' : '已取消所有组的自动跟随；已有人员及排队投递保持不变。');
        }).fail(function(xhr) { $('#autoGroupsFeedback').text(xhr.responseJSON && xhr.responseJSON.message || '保存失败，请重试'); }).always(function () {button.prop('disabled',false);});
    });
});
