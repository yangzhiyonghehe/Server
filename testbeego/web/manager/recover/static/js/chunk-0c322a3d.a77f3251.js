(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-0c322a3d"],{"4bf8b":function(t,n,e){"use strict";e.d(n,"b",(function(){return a})),e.d(n,"g",(function(){return o})),e.d(n,"c",(function(){return i})),e.d(n,"e",(function(){return c})),e.d(n,"d",(function(){return u})),e.d(n,"f",(function(){return f})),e.d(n,"a",(function(){return s}));var r=e("b775");function a(t){return Object(r["a"])({url:"/flow/form",method:"get",params:{id:t}})}function o(t){return Object(r["a"])({url:"/flow/saveform",method:"post",data:t})}function i(t){return Object(r["a"])({url:"/flow/list",method:"get",params:t})}function c(t){return Object(r["a"])({url:"/flow/saveApp",method:"post",data:t})}function u(t){return Object(r["a"])({url:"/flow/delete",method:"post",data:t})}function f(t){return Object(r["a"])({url:"/flow/saveFlow",method:"post",data:t})}function s(t){return Object(r["a"])({url:"/flow/fetch",method:"get",params:{id:t}})}},"70a5":function(t,n,e){"use strict";var r=e("998a"),a=e.n(r);a.a},"998a":function(t,n,e){},a1b7:function(t,n,e){"use strict";var r=e("acb8"),a=e.n(r);a.a},acb8:function(t,n,e){},d4fa:function(t,n,e){"use strict";e.r(n);var r=function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("div",{staticClass:"wrap"},[e("fm-making-form",{ref:"makingForm",staticStyle:{height:"calc(100vh - 84px)"},attrs:{preview:""}},[e("template",{slot:"action"},[e("el-button",{attrs:{type:"text",size:"medium",icon:"el-icon-finished"},on:{click:t.handleSave}},[t._v(t._s(t.$t("workflow.form.saveDesign")))])],1)],2)],1)},a=[],o=e("4bf8b"),i={name:"WorkflowForm",data:function(){return{id:void 0}},mounted:function(){var t=this.$route.params;this.id=t.id,this.fetch()},methods:{handleSave:function(){var t=this,n=JSON.stringify(this.$refs.makingForm.getJSON());Object(o["g"])({id:this.id,form:n}).then((function(n){t.$confirm(t.$t("workflow.form.saveSuccessMsg"),t.$t("table.confirmTitle"),{confirmButtonText:t.$t("table.confirm"),cancelButtonText:t.$t("table.cancel"),type:"warning"}).then((function(){t.$router.push({name:"Flow",params:{id:t.id}})})).catch((function(){return!1}))}))},fetch:function(){var t=this;Object(o["b"])(this.id).then((function(n){var e=JSON.parse(n.data);t.$refs.makingForm.setJSON(e)}))}}},c=i,u=(e("a1b7"),e("70a5"),e("2877")),f=Object(u["a"])(c,r,a,!1,null,"69143a4f",null);n["default"]=f.exports}}]);