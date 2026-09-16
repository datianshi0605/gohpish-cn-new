var templates = []
var icons = {
    "application/vnd.ms-excel": "fa-file-excel-o",
    "text/plain": "fa-file-text-o",
    "image/gif": "fa-file-image-o",
    "image/png": "fa-file-image-o",
    "application/pdf": "fa-file-pdf-o",
    "application/x-zip-compressed": "fa-file-archive-o",
    "application/x-gzip": "fa-file-archive-o",
    "application/vnd.openxmlformats-officedocument.presentationml.presentation": "fa-file-powerpoint-o",
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": "fa-file-word-o",
    "application/octet-stream": "fa-file-o",
    "application/x-msdownload": "fa-file-o"
}

var qrPreviewImageSrc = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAUAAAAFAAQMAAAD3XjfpAAAABlBMVEX///8AAABVwtN+AAACY0lEQVR42uyavbEkKwyFD4WBSQiEQmjvhkYohICJQaFXEjDbfaebmV23JWNra/qzdPVzJAE1NTU1tXsjsQbAVgCueCLKISOMLz8KfgfyP6bBUnMVjoov8Bkhh8Rf/ns6aIi6abZZYq4CvnjKQEiRKCn4lyA1OA5IiMMDZQX/CWyWqDoiKvD04S+j4AU4CoClZqu43NMskdeVQsEbcPQZ22xz1VVXfPEl5JBvGtLDwGm2QfzoJBwlIK9NwWuQ85rlzEjrKl2mgOsj/54U/BJEN910Tmv+IBWSwYyQEI/1UcE9SEw2CV1HVLldjwKQYorH/FdwB0rgdvZ2c6IgxeHsccT3CH8ciDWm8CDnpprhrM7MxR8FvwOpm+lwHlEcF0j+masj68JDY1dwC1IHDJdHDlzHQSt5TZnT+pjXCm5B02G6/NeJ0+f6YYCnSqHgFqS16HKjjooOR8iCnefC54FDzrAqpCWvix9TSrrQPQregZzYhljP1OluWXLhXfcouAP7zGqWhVIfpV8j/GozCu5BVtcwc+CT+kgjcFlAHj2u4B7EYdyr7rUtlAKQTuuwR4Jm5PXaKmDqwjENn6c4BTcghrx+bQuxjnH4ldcKbkGz2gzNG4kvCGvHFRX8Elxb1zXvyfVzrh9SPOoeBXfg68Y+K8DYcsmRhMGLtwrPAucNaSxnRheWpfTowm9PEBS8AddtmJ1+3HJJYsfzMlXBz2Bjf68tF0euLF2vj/EKbsDx6ONw/Dzlv4IfwD9vkLC2ruOxQqSrQvo0cN7YqVlWhXIhwbyxn3WPgjtQTU1N7Tn2fwAAAP//6EclY07jIJ8AAAAASUVORK5CYII="
var qrBlockSnippet = '<!-- Gophish QR Code -->' +
    '<div class="qr-section" style="text-align:center;margin:24px auto;">' +
    '<img data-gophish-qr="true" data-gophish-mode="html" data-gophish-src="{{.QRCodeHTML}}" src="' + qrPreviewImageSrc + '" width="220" height="220" alt="信息核验二维码" style="display:block;width:220px;height:220px;margin:0 auto;border:1px solid #e5e5e5;padding:8px;background:#fff;">' +
    '<div class="qr-caption" style="font-size:14px;color:#555;margin-top:12px;">请使用手机扫码进入信息核验平台</div>' +
    '</div>' +
    '<p style="text-align:center;margin:28px 0 0;">' +
    '<a data-gophish-qr-link="true" href="{{.URL}}" style="display:inline-block;background:#1677ff;color:#ffffff;text-decoration:none;padding:14px 24px;border-radius:6px;font-size:16px;font-weight:bold;">无法扫码时，点击进入信息核验平台</a>' +
    '</p>' +
    '<!-- /Gophish QR Code -->'
var dataTableChinese = {
    sEmptyTable: "暂无数据",
    sZeroRecords: "没有找到匹配记录",
    sInfo: "显示第 _START_ 到 _END_ 条，共 _TOTAL_ 条",
    sInfoEmpty: "显示第 0 到 0 条，共 0 条",
    sInfoFiltered: "（从 _MAX_ 条记录中过滤）",
    sLengthMenu: "每页显示 _MENU_ 条",
    sLoadingRecords: "正在加载...",
    sProcessing: "处理中...",
    sSearch: "搜索：",
    oPaginate: {
        sFirst: "首页",
        sLast: "末页",
        sNext: "下一页",
        sPrevious: "上一页"
    }
}
var attachmentTableOptions = {
    destroy: true,
    paging: false,
    searching: false,
    info: false,
    lengthChange: false,
    oLanguage: $.extend({}, dataTableChinese, {
        sEmptyTable: ""
    }),
    "order": [
        [1, "asc"]
    ],
    columnDefs: [{
        orderable: false,
        targets: "no-sort"
    }, {
        sClass: "datatable_hidden",
        targets: [3, 4]
    }]
}
var qrCodeAutoTargets = [
    /(<h[1-6][^>]*>\s*(?:二[、.．]\s*)?信息核验要求\s*<\/h[1-6]>\s*<p[\s\S]*?<\/p>)/i,
    /(<h[1-6][^>]*>\s*(?:二[、.．]\s*)?核验要求\s*<\/h[1-6]>\s*<p[\s\S]*?<\/p>)/i,
    /(<h[1-6][^>]*>\s*(?:二[、.．]\s*)?安全确认\s*<\/h[1-6]>\s*<p[\s\S]*?<\/p>)/i
]

function hasQrCode(html) {
    return html.indexOf("{{.QRCodeURL}}") != -1 || html.indexOf("{{.QRCodeHTML}}") != -1 || html.indexOf("{{.QRCode}}") != -1 || html.indexOf("data-gophish-qr") != -1
}

function normalizeQrPlaceholders(html) {
    return html.replace(/<img([^>]*data-gophish-qr=["']true["'][^>]*)>/ig, function(match) {
        if (match.match(/\sdata-gophish-mode=["']html["']/i) || match.match(/\sdata-gophish-src=["']{{\.QRCodeHTML}}["']/i)) {
            return "{{.QRCodeHTML}}"
        }
        var normalized = match
        if (normalized.match(/\ssrc=["'][^"']*["']/i)) {
            normalized = normalized.replace(/\ssrc=["'][^"']*["']/i, ' src="{{.QRCodeURL}}"')
        } else {
            normalized = normalized.replace("<img", '<img src="{{.QRCodeURL}}"')
        }
        normalized = normalized.replace(/\sdata-gophish-src=["'][^"']*["']/ig, "")
        return normalized
    })
}

function looksLikeQrImage(attrs) {
    return attrs.indexOf("{{.QRCodeURL}}") != -1 ||
        attrs.indexOf("data-gophish-qr") != -1 ||
        attrs.indexOf("data-gophish-src") != -1 ||
        /alt=["'][^"']*(二维码|QR Code|qr code)[^"']*["']/i.test(attrs) ||
        /class=["'][^"']*qr[^"']*["']/i.test(attrs)
}

function prepareQrPreviewPlaceholders(html) {
    if (!html || !qrPreviewImageSrc) {
        return html
    }
    html = html.replace(/{{\.QRCodeHTML}}/g, '<img data-gophish-qr="true" data-gophish-mode="html" data-gophish-src="{{.QRCodeHTML}}" src="' + qrPreviewImageSrc + '" width="220" height="220" alt="信息核验二维码" style="display:block;width:220px;height:220px;margin:0 auto;border:1px solid #e5e5e5;padding:8px;background:#fff;">')
    return html.replace(/<img([^>]*)>/ig, function(match, attrs) {
        if (!looksLikeQrImage(attrs)) {
            return match
        }
        var preview = attrs
        if (!preview.match(/\sdata-gophish-qr=["']true["']/i)) {
            preview += ' data-gophish-qr="true"'
        }
        if (!preview.match(/\sdata-gophish-src=["'][^"']*["']/i)) {
            preview += ' data-gophish-src="{{.QRCodeHTML}}" data-gophish-mode="html"'
        }
        if (preview.match(/\ssrc=["'][^"']*["']/i)) {
            preview = preview.replace(/\ssrc=["'][^"']*["']/i, ' src="' + qrPreviewImageSrc + '"')
        } else {
            preview = ' src="' + qrPreviewImageSrc + '"' + preview
        }
        return "<img" + preview + ">"
    })
}

function removeManagedQrBlock(html) {
    return html.replace(/<!-- Gophish QR Code -->[\s\S]*?<!-- \/Gophish QR Code -->/g, "")
}

function insertQrBlock(html) {
    var placeholder = "<!-- Gophish QR Code Here -->"
    if (html.indexOf(placeholder) != -1) {
        return html.replace(placeholder, qrBlockSnippet)
    }
    for (var i = 0; i < qrCodeAutoTargets.length; i++) {
        if (qrCodeAutoTargets[i].test(html)) {
            return html.replace(qrCodeAutoTargets[i], "$1" + qrBlockSnippet)
        }
    }
    if (html.indexOf("</body>") != -1) {
        return html.replace("</body>", qrBlockSnippet + "</body>")
    }
    return html + qrBlockSnippet
}

function syncQrCheckbox() {
    if (!CKEDITOR.instances["html_editor"]) {
        return
    }
    var editor = CKEDITOR.instances["html_editor"]
    var updateQrBlock = function() {
        var html = editor.getData()
        if ($("#use_qr_checkbox").prop("checked")) {
            if (!hasQrCode(html)) {
                editor.setData(insertQrBlock(html))
            } else if (html.indexOf("<!-- Gophish QR Code -->") != -1) {
                editor.setData(insertQrBlock(removeManagedQrBlock(html)))
            }
        } else {
            html = removeManagedQrBlock(html)
            editor.setData(html)
        }
        $('.nav-tabs a[href="#html"]').click()
    }
    if (editor.mode == "source") {
        editor.once("mode", updateQrBlock)
        editor.setMode("wysiwyg")
        return
    }
    updateQrBlock()
}

// Save attempts to POST to /templates/
function save(idx) {
    var template = {
        attachments: []
    }
    template.name = $("#name").val()
    template.subject = $("#subject").val()
    template.envelope_sender = $("#envelope-sender").val()
    template.html = CKEDITOR.instances["html_editor"].getData();
    // Fix the URL Scheme added by CKEditor (until we can remove it from the plugin)
    template.html = template.html.replace(/https?:\/\/{{\.URL}}/gi, "{{.URL}}")
    template.html = normalizeQrPlaceholders(template.html)
    // If the "Add Tracker Image" checkbox is checked, add the tracker
    if ($("#use_tracker_checkbox").prop("checked")) {
        if (template.html.indexOf("{{.Tracker}}") == -1 &&
            template.html.indexOf("{{.TrackingUrl}}") == -1) {
            template.html = template.html.replace("</body>", "{{.Tracker}}</body>")
        }
    } else {
        // Otherwise, remove the tracker
        template.html = template.html.replace("{{.Tracker}}</body>", "</body>")
    }
    template.text = $("#text_editor").val()
    // Add the attachments
    $.each($("#attachmentsTable").DataTable().rows().data(), function (i, target) {
        template.attachments.push({
            name: unescapeHtml(target[1]),
            content: target[3],
            type: target[4],
        })
    })

    if (idx != -1) {
        template.id = templates[idx].id
        api.templateId.put(template)
            .success(function (data) {
                successFlash("模板已保存")
                load()
                dismiss()
            })
            .error(function (data) {
                modalError(data.responseJSON.message)
            })
    } else {
        // Submit the template
        api.templates.post(template)
            .success(function (data) {
                successFlash("模板已创建")
                load()
                dismiss()
            })
            .error(function (data) {
                modalError(data.responseJSON.message)
            })
    }
}

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#attachmentsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#subject").val("")
    $("#text_editor").val("")
    $("#html_editor").val("")
    $("#use_tracker_checkbox").prop("checked", true)
    $("#use_qr_checkbox").prop("checked", false)
    $("#modal").modal('hide')
}

var deleteTemplate = function (idx) {
    Swal.fire({
        title: "确认删除？",
        text: "删除后无法恢复。",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete " + escapeHtml(templates[idx].name),
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function () {
            return new Promise(function (resolve, reject) {
                api.templateId.delete(templates[idx].id)
                    .success(function (msg) {
                        resolve()
                    })
                    .error(function (data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function (result) {
        if(result.value) {
            Swal.fire(
                '模板已删除',
                '该模板已经删除。',
                'success'
            );
        }
        $('button:contains("OK")').on('click', function () {
            location.reload()
        })
    })
}

function deleteTemplate(idx) {
    if (confirm("Delete " + templates[idx].name + "?")) {
        api.templateId.delete(templates[idx].id)
            .success(function (data) {
                successFlash(data.message)
                load()
            })
    }
}

function attach(files) {
    attachmentsTable = $("#attachmentsTable").DataTable(attachmentTableOptions);
    $.each(files, function (i, file) {
        var reader = new FileReader();
        /* Make this a datatable */
        reader.onload = function (e) {
            var icon = icons[file.type] || "fa-file-o"
            // Add the record to the modal
            attachmentsTable.row.add([
                '<i class="fa ' + icon + '"></i>',
                escapeHtml(file.name),
                '<span class="remove-row"><i class="fa fa-trash-o"></i></span>',
                reader.result.split(",")[1],
                file.type || "application/octet-stream"
            ]).draw()
        }
        reader.onerror = function (e) {
            console.log(e)
        }
        reader.readAsDataURL(file)
    })
}

function edit(idx) {
    $("#modalSubmit").unbind('click').click(function () {
        save(idx)
    })
    $("#attachmentUpload").unbind('click').click(function () {
        this.value = null
    })
    $("#html_editor").ckeditor()
    setupAutocomplete(CKEDITOR.instances["html_editor"])
    $("#attachmentsTable").show()
    attachmentsTable = $('#attachmentsTable').DataTable(attachmentTableOptions);
    var template = {
        attachments: []
    }
    if (idx != -1) {
        $("#templateModalLabel").text("编辑模板")
        template = templates[idx]
        $("#name").val(template.name)
        $("#subject").val(template.subject)
        $("#envelope-sender").val(template.envelope_sender)
        $("#html_editor").val(prepareQrPreviewPlaceholders(template.html))
        $("#text_editor").val(template.text)
        attachmentRows = []
        $.each(template.attachments, function (i, file) {
            var icon = icons[file.type] || "fa-file-o"
            // Add the record to the modal
            attachmentRows.push([
                '<i class="fa ' + icon + '"></i>',
                escapeHtml(file.name),
                '<span class="remove-row"><i class="fa fa-trash-o"></i></span>',
                file.content,
                file.type || "application/octet-stream"
            ])
        })
        attachmentsTable.rows.add(attachmentRows).draw()
        if (template.html.indexOf("{{.Tracker}}") != -1) {
            $("#use_tracker_checkbox").prop("checked", true)
        } else {
            $("#use_tracker_checkbox").prop("checked", false)
        }
        $("#use_qr_checkbox").prop("checked", hasQrCode(template.html))
    } else {
        $("#templateModalLabel").text("新建模板")
        $("#use_tracker_checkbox").prop("checked", true)
        $("#use_qr_checkbox").prop("checked", false)
    }
    // Handle Deletion
    $("#attachmentsTable").unbind('click').on("click", "span>i.fa-trash-o", function () {
        attachmentsTable.row($(this).parents('tr'))
            .remove()
            .draw();
    })
}

function copy(idx) {
    $("#modalSubmit").unbind('click').click(function () {
        save(-1)
    })
    $("#attachmentUpload").unbind('click').click(function () {
        this.value = null
    })
    $("#html_editor").ckeditor()
    $("#attachmentsTable").show()
    attachmentsTable = $('#attachmentsTable').DataTable(attachmentTableOptions);
    var template = {
        attachments: []
    }
    template = templates[idx]
    $("#name").val("复制 " + template.name)
    $("#subject").val(template.subject)
    $("#envelope-sender").val(template.envelope_sender)
    $("#html_editor").val(prepareQrPreviewPlaceholders(template.html))
    $("#text_editor").val(template.text)
    $.each(template.attachments, function (i, file) {
        var icon = icons[file.type] || "fa-file-o"
        // Add the record to the modal
        attachmentsTable.row.add([
            '<i class="fa ' + icon + '"></i>',
            escapeHtml(file.name),
            '<span class="remove-row"><i class="fa fa-trash-o"></i></span>',
            file.content,
            file.type || "application/octet-stream"
        ]).draw()
    })
    // Handle Deletion
    $("#attachmentsTable").unbind('click').on("click", "span>i.fa-trash-o", function () {
        attachmentsTable.row($(this).parents('tr'))
            .remove()
            .draw();
    })
    if (template.html.indexOf("{{.Tracker}}") != -1) {
        $("#use_tracker_checkbox").prop("checked", true)
    } else {
        $("#use_tracker_checkbox").prop("checked", false)
    }
    $("#use_qr_checkbox").prop("checked", hasQrCode(template.html))
}

function importEmail() {
    raw = $("#email_content").val()
    convert_links = $("#convert_links_checkbox").prop("checked")
    if (!raw) {
        modalError("请先填写邮件内容")
    } else {
        api.import_email({
                content: raw,
                convert_links: convert_links
            })
            .success(function (data) {
                $("#text_editor").val(data.text)
                $("#html_editor").val(prepareQrPreviewPlaceholders(data.html))
                $("#subject").val(data.subject)
                // If the HTML is provided, let's open that view in the editor
                if (data.html) {
                    CKEDITOR.instances["html_editor"].setMode('wysiwyg')
                    $('.nav-tabs a[href="#html"]').click()
                }
                $("#importEmailModal").modal("hide")
            })
            .error(function (data) {
                modalError(data.responseJSON.message)
            })
    }
}

function load() {
    $("#templateTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.templates.get()
        .success(function (ts) {
            templates = ts
            $("#loading").hide()
            if (templates.length > 0) {
                $("#templateTable").show()
                templateTable = $("#templateTable").DataTable({
                    destroy: true,
                    oLanguage: dataTableChinese,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                templateTable.clear()
                templateRows = []
                $.each(templates, function (i, template) {
                    templateRows.push([
                        escapeHtml(template.name),
                        moment(template.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><span data-toggle='modal' data-backdrop='static' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Template' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button></span>\
		    <span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Template' onclick='copy(" + i + ")'>\
                    <i class='fa fa-copy'></i>\
                    </button></span>\
                    <button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Template' onclick='deleteTemplate(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ])
                })
                templateTable.rows.add(templateRows).draw()
                $('[data-toggle="tooltip"]').tooltip()
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function () {
            $("#loading").hide()
            errorFlash("Error fetching templates")
        })
}

$(document).ready(function () {
    // Setup multiple modals
    // Code based on http://miles-by-motorcycle.com/static/bootstrap-modal/index.html
    $('.modal').on('hidden.bs.modal', function (event) {
        $(this).removeClass('fv-modal-stack');
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') - 1);
    });
    $('.modal').on('shown.bs.modal', function (event) {
        // Keep track of the number of open modals
        if (typeof ($('body').data('fv_open_modals')) == 'undefined') {
            $('body').data('fv_open_modals', 0);
        }
        // if the z-index of this modal has been set, ignore.
        if ($(this).hasClass('fv-modal-stack')) {
            return;
        }
        $(this).addClass('fv-modal-stack');
        // Increment the number of open modals
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') + 1);
        // Setup the appropriate z-index
        $(this).css('z-index', 1040 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('.fv-modal-stack').css('z-index', 1039 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('fv-modal-stack').addClass('fv-modal-stack');
    });
    $.fn.modal.Constructor.prototype.enforceFocus = function () {
        $(document)
            .off('focusin.bs.modal') // guard against infinite focus loop
            .on('focusin.bs.modal', $.proxy(function (e) {
                if (
                    this.$element[0] !== e.target && !this.$element.has(e.target).length
                    // CKEditor compatibility fix start.
                    &&
                    !$(e.target).closest('.cke_dialog, .cke').length
                    // CKEditor compatibility fix end.
                ) {
                    this.$element.trigger('focus');
                }
            }, this));
    };
    // Scrollbar fix - https://stackoverflow.com/questions/19305821/multiple-modals-overlay
    $(document).on('hidden.bs.modal', '.modal', function () {
        $('.modal:visible').length && $(document.body).addClass('modal-open');
    });
    $('#modal').on('hidden.bs.modal', function (event) {
        dismiss()
    });
    $("#importEmailModal").on('hidden.bs.modal', function (event) {
        $("#email_content").val("")
    })
    $("#use_qr_checkbox").change(syncQrCheckbox)
    CKEDITOR.on('dialogDefinition', function (ev) {
        // Take the dialog name and its definition from the event data.
        var dialogName = ev.data.name;
        var dialogDefinition = ev.data.definition;

        // Check if the definition is from the dialog window you are interested in (the "Link" dialog window).
        if (dialogName == 'link') {
            dialogDefinition.minWidth = 500
            dialogDefinition.minHeight = 100

            // Remove the linkType field
            var infoTab = dialogDefinition.getContents('info');
            infoTab.get('linkType').hidden = true;
        }
    });
    load()

})
