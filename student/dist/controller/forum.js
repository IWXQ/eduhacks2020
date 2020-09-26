/** layuiAdmin.pro-v1.4.0 LPPL License By https://www.layui.com/admin/ */
;layui.define(["table", "form"], function (e) {
    var t = (layui.$, layui.admin), i = layui.view, l = layui.table, r = layui.form;
    l.render({
        elem: "#LAY-app-forum-list",
        url: "./json/forum/upload.js",
        cols: [[{type: "checkbox", fixed: "left"}, {field: "id", width: 100, title: "ID", sort: !0}, {
            field: "poster",
            title: "发帖人"
        }, {field: "avatar", title: "头像", width: 100, templet: "#imgTpl"}, {
            field: "content",
            title: "发帖内容"
        }, {field: "posttime", title: "发帖时间", sort: !0}, {
            field: "top",
            title: "置顶",
            templet: "#buttonTpl",
            minWidth: 80,
            align: "center"
        }, {title: "操作", width: 150, align: "center", fixed: "right", toolbar: "#table-forum-list"}]],
        page: !0,
        limit: 10,
        limits: [10, 15, 20, 25, 30],
        text: "对不起，加载出现异常！"
    }), l.on("tool(LAY-app-forum-list)", function (e) {
        var l = e.data;
        "del" === e.event ? layer.confirm("确定删除此条帖子？", function (t) {
            e.del(), layer.close(t)
        }) : "edit" === e.event && t.popup({
            title: "编辑帖子",
            area: ["550px", "450px"],
            id: "LAY-popup-forum-edit",
            resize: !1,
            success: function (e, t) {
                i(this.id).render("app/forum/listform", l).done(function () {
                    r.render(null, "layuiadmin-form-list"), r.on("submit(layuiadmin-app-forum-submit)", function (e) {
                        e.field;
                        layui.table.reload("LAY-app-forum-list"), layer.close(t)
                    })
                })
            }
        })
    }), l.render({
        elem: "#LAY-app-forumreply-list",
        url: "./json/forum/replys.js",
        cols: [[{type: "checkbox", fixed: "left"}, {field: "id", width: 100, title: "ID", sort: !0}, {
            field: "replyer",
            title: "回帖人"
        }, {field: "cardid", title: "回帖ID", sort: !0}, {
            field: "avatar",
            title: "头像",
            width: 100,
            templet: "#imgTpl"
        }, {field: "content", title: "回帖内容", width: 200}, {field: "replytime", title: "回帖时间", sort: !0}, {
            title: "操作",
            width: 150,
            align: "center",
            fixed: "right",
            toolbar: "#table-forum-replys"
        }]],
        page: !0,
        limit: 10,
        limits: [10, 15, 20, 25, 30],
        text: "对不起，加载出现异常！"
    }), l.on("tool(LAY-app-forumreply-list)", function (e) {
        var l = e.data;
        "del" === e.event ? layer.confirm("确定删除此条评论？", function (t) {
            e.del(), layer.close(t)
        }) : "edit" === e.event && t.popup({
            title: "编辑回帖",
            area: ["550px", "400px"],
            id: "LAY-popup-forum-edit",
            resize: !1,
            success: function (e, t) {
                i(this.id).render("app/forum/replysform", l).done(function () {
                    r.render(null, "layuiadmin-app-forum-reply"), r.on("submit(layuiadmin-app-forumreply-submit)", function (e) {
                        e.field;
                        layui.table.reload("LAY-app-forumreply-list"), layer.close(t)
                    })
                })
            }
        })
    }), e("forum", {})
});