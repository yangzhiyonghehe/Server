(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["chunk-34d14b62"],{"09f4":function(t,e,a){"use strict";a.d(e,"a",(function(){return r})),Math.easeInOutQuad=function(t,e,a,i){return t/=i/2,t<1?a/2*t*t+e:(t--,-a/2*(t*(t-2)-1)+e)};var i=function(){return window.requestAnimationFrame||window.webkitRequestAnimationFrame||window.mozRequestAnimationFrame||function(t){window.setTimeout(t,1e3/60)}}();function n(t){document.documentElement.scrollTop=t,document.body.parentNode.scrollTop=t,document.body.scrollTop=t}function o(){return document.documentElement.scrollTop||document.body.parentNode.scrollTop||document.body.scrollTop}function r(t,e,a){var r=o(),l=t-r,s=20,u=0;e="undefined"===typeof e?500:e;var c=function t(){u+=s;var o=Math.easeInOutQuad(u,r,l,e);n(o),u<e?i(t):a&&"function"===typeof a&&a()};c()}},"4bf8b":function(t,e,a){"use strict";a.d(e,"b",(function(){return n})),a.d(e,"g",(function(){return o})),a.d(e,"c",(function(){return r})),a.d(e,"e",(function(){return l})),a.d(e,"d",(function(){return s})),a.d(e,"f",(function(){return u})),a.d(e,"a",(function(){return c}));var i=a("b775");function n(t){return Object(i["a"])({url:"/flow/form",method:"get",params:{id:t}})}function o(t){return Object(i["a"])({url:"/flow/saveform",method:"post",data:t})}function r(t){return Object(i["a"])({url:"/flow/list",method:"get",params:t})}function l(t){return Object(i["a"])({url:"/flow/saveApp",method:"post",data:t})}function s(t){return Object(i["a"])({url:"/flow/delete",method:"post",data:t})}function u(t){return Object(i["a"])({url:"/flow/saveFlow",method:"post",data:t})}function c(t){return Object(i["a"])({url:"/flow/fetch",method:"get",params:{id:t}})}},"5e0d":function(t,e,a){"use strict";var i=a("6044"),n=a.n(i);n.a},6044:function(t,e,a){},6724:function(t,e,a){"use strict";a("8d41");var i="@@wavesContext";function n(t,e){function a(a){var i=Object.assign({},e.value),n=Object.assign({ele:t,type:"hit",color:"rgba(0, 0, 0, 0.15)"},i),o=n.ele;if(o){o.style.position="relative",o.style.overflow="hidden";var r=o.getBoundingClientRect(),l=o.querySelector(".waves-ripple");switch(l?l.className="waves-ripple":(l=document.createElement("span"),l.className="waves-ripple",l.style.height=l.style.width=Math.max(r.width,r.height)+"px",o.appendChild(l)),n.type){case"center":l.style.top=r.height/2-l.offsetHeight/2+"px",l.style.left=r.width/2-l.offsetWidth/2+"px";break;default:l.style.top=(a.pageY-r.top-l.offsetHeight/2-document.documentElement.scrollTop||document.body.scrollTop)+"px",l.style.left=(a.pageX-r.left-l.offsetWidth/2-document.documentElement.scrollLeft||document.body.scrollLeft)+"px"}return l.style.backgroundColor=n.color,l.className="waves-ripple z-active",!1}}return t[i]?t[i].removeHandle=a:t[i]={removeHandle:a},a}var o={bind:function(t,e){t.addEventListener("click",n(t,e),!1)},update:function(t,e){t.removeEventListener("click",t[i].removeHandle,!1),t.addEventListener("click",n(t,e),!1)},unbind:function(t){t.removeEventListener("click",t[i].removeHandle,!1),t[i]=null,delete t[i]}},r=function(t){t.directive("waves",o)};window.Vue&&(window.waves=o,Vue.use(r)),o.install=r;e["a"]=o},"8d41":function(t,e,a){},b36a:function(t,e,a){"use strict";a.r(e);var i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"app-container"},[a("div",{staticClass:"filter-container"},[a("el-button",{staticClass:"filter-item",staticStyle:{"margin-left":"10px"},attrs:{type:"primary",icon:"el-icon-edit"},on:{click:t.handleCreate}},[t._v(t._s(t.$t("workflow.approval.addBtn")))])],1),t._v(" "),a("el-table",{directives:[{name:"loading",rawName:"v-loading",value:t.listLoading,expression:"listLoading"}],key:0,staticStyle:{width:"100%"},attrs:{data:t.list,border:"",fit:"","highlight-current-row":""},on:{"sort-change":t.sortChange}},[a("el-table-column",{attrs:{label:"ID",prop:"id",sortable:"custom",align:"center",width:"80"},scopedSlots:t._u([{key:"default",fn:function(e){var i=e.row;return[a("span",[t._v(t._s(i.id))])]}}])}),t._v(" "),a("el-table-column",{attrs:{label:t.$t("workflow.approval.appName"),"min-width":"100"},scopedSlots:t._u([{key:"default",fn:function(e){var i=e.row;return[a("span",[t._v(t._s(i.name))])]}}])}),t._v(" "),a("el-table-column",{attrs:{label:t.$t("table.sort"),"min-width":"100"},scopedSlots:t._u([{key:"default",fn:function(e){var i=e.row;return[a("span",[t._v(t._s(i.sort))])]}}])}),t._v(" "),a("el-table-column",{attrs:{label:t.$t("table.status"),width:"100"},scopedSlots:t._u([{key:"default",fn:function(e){var i=e.row;return[a("el-tag",{attrs:{type:t.statusTagMap[i.status]}},[t._v(t._s(t.statusMap[i.status]))])]}}])}),t._v(" "),a("el-table-column",{attrs:{label:t.$t("table.actions"),align:"center",width:"350","class-name":"small-padding fixed-width"},scopedSlots:t._u([{key:"default",fn:function(e){var i=e.row;return[a("el-button",{staticStyle:{width:"auto"},attrs:{type:"primary",size:"mini"},on:{click:function(e){return t.handleApply(i)}}},[t._v(t._s(t.$t("workflow.approval.record")))]),t._v(" "),a("el-button-group",[a("el-button",{attrs:{type:"primary",size:"mini"},on:{click:function(e){return t.handleUpdate(i)}}},[t._v(t._s(t.$t("table.edit")))]),t._v(" "),a("el-button",{directives:[{name:"waves",rawName:"v-waves"}],staticStyle:{width:"auto"},attrs:{type:"primary",size:"mini"},on:{click:function(e){return t.handleDesign(i.id)}}},[t._v(t._s(t.$t("workflow.approval.design")))]),t._v(" "),"deleted"!=i.status?a("el-button",{attrs:{size:"mini",type:"danger"},on:{click:function(e){return t.handleDelete(i)}}},[t._v(t._s(t.$t("table.delete")))]):t._e()],1)]}}])})],1),t._v(" "),a("pagination",{directives:[{name:"show",rawName:"v-show",value:t.total>0,expression:"total > 0"}],attrs:{total:t.total,page:t.listQuery.page,limit:t.listQuery.limit},on:{"update:page":function(e){return t.$set(t.listQuery,"page",e)},"update:limit":function(e){return t.$set(t.listQuery,"limit",e)},pagination:t.getList}}),t._v(" "),a("el-dialog",{attrs:{title:t.$t("workflow.dialog1.name"),visible:t.dialogFormVisible},on:{"update:visible":function(e){t.dialogFormVisible=e}}},[a("div",{staticClass:"form-dialog"},[a("el-form",{ref:"dataForm",staticClass:"form-wrap",staticStyle:{width:"400px","margin-left":"50px"},attrs:{rules:t.rules,model:t.temp,"label-position":"left","label-width":"100px"}},[a("el-form-item",{attrs:{label:t.$t("workflow.dialog1.appName"),prop:"name"}},[a("el-input",{model:{value:t.temp.name,callback:function(e){t.$set(t.temp,"name",e)},expression:"temp.name"}})],1),t._v(" "),a("el-form-item",{attrs:{label:t.$t("table.sort"),prop:"sort"}},[a("el-input",{model:{value:t.temp.sort,callback:function(e){t.$set(t.temp,"sort",e)},expression:"temp.sort"}})],1),t._v(" "),a("el-form-item",{attrs:{label:t.$t("workflow.dialog1.appStatus")}},[a("el-switch",{attrs:{"active-value":1,"inactive-value":0,"active-text":t.$t("workflow.dialog1.enable")},model:{value:t.temp.status,callback:function(e){t.$set(t.temp,"status",e)},expression:"temp.status"}})],1)],1)],1),t._v(" "),a("div",{staticClass:"dialog-footer",attrs:{slot:"footer"},slot:"footer"},[a("el-button",{on:{click:function(e){t.dialogFormVisible=!1}}},[t._v(t._s(t.$t("table.cancel")))]),t._v(" "),a("el-button",{attrs:{type:"primary"},on:{click:function(e){"create"===t.dialogStatus?t.createData():t.updateData()}}},[t._v(t._s(t.$t("workflow.dialog1.next"))),a("i",{staticClass:"el-icon-right el-icon--right"})])],1)])],1)},n=[],o=(a("ac4d"),a("8a81"),a("ac6a"),a("4bf8b")),r=a("6724"),l=a("333d"),s={name:"WorkflowIndex",components:{Pagination:l["a"]},directives:{waves:r["a"]},data:function(){return{statusMap:{0:this.$t("workflow.dialog1.disabled"),1:this.$t("workflow.dialog1.enabled")},statusTagMap:{0:"info",1:"success"},dialogFormVisible:!1,list:null,total:0,listLoading:!0,listQuery:{page:1,limit:20,name:void 0,sort:"+id"},temp:{},dialogStatus:"",rules:{name:[{required:!0,message:this.$t("workflow.required.name"),trigger:"blur"}],sort:[{required:!0,message:this.$t("workflow.required.sort"),trigger:"blur"}]}}},created:function(){this.resetTemp(),this.getList()},methods:{handleDesign:function(t){this.$router.push({name:"Flow",params:{id:t}})},getList:function(){var t=this;this.listLoading=!0,Object(o["c"])(this.listQuery).then((function(e){t.list=e.data.items,t.total=e.data.total,t.listLoading=!1}))},handleFilter:function(){this.listQuery.page=1,this.getList()},sortChange:function(t){var e=t.prop,a=t.order;"id"===e&&this.sortByID(a)},sortByID:function(t){this.listQuery.sort="ascending"===t?"+id":"-id",this.handleFilter()},resetTemp:function(){this.temp={id:void 0,name:"",sort:0,status:0}},handleCreate:function(){this.resetTemp(),this.dialogStatus="create",this.dialogFormVisible=!0},createData:function(){var t=this;this.$refs["dataForm"].validate((function(e){e&&Object(o["e"])(t.temp).then((function(e){t.temp=e.data,t.list.unshift(t.temp),t.$router.push({name:"WorkflowForm",params:{id:t.temp.id}})}))}))},handleApply:function(t){this.$router.push({name:"WorkflowApproval",params:{id:t.id}})},handleUpdate:function(t){this.temp=Object.assign({},t),this.dialogFormVisible=!0},updateData:function(){var t=this;this.$refs["dataForm"].validate((function(e){e&&Object(o["e"])(t.temp).then((function(e){t.temp=e.data;var a=!0,i=!1,n=void 0;try{for(var o,r=t.list[Symbol.iterator]();!(a=(o=r.next()).done);a=!0){var l=o.value;if(l.id===t.temp.id){var s=t.list.indexOf(l);t.list.splice(s,1,t.temp);break}}}catch(u){i=!0,n=u}finally{try{a||null==r.return||r.return()}finally{if(i)throw n}}t.$router.push({name:"WorkflowForm",params:{id:t.temp.id}})}))}))},handleDelete:function(t){var e=this;this.$confirm(this.$t("table.confirmBody"),this.$t("table.confirmTitle"),{confirmButtonText:this.$t("table.confirm"),cancelButtonText:this.$t("table.cancel"),type:"warning"}).then((function(){Object(o["d"])(t).then((function(){e.$notify({title:e.$t("table.successTitle"),message:e.$t("table.deleteSuccess"),type:"success",duration:2e3});var a=e.list.indexOf(t);e.list.splice(a,1)}))}))}}},u=s,c=(a("5e0d"),a("2877")),d=Object(c["a"])(u,i,n,!1,null,"18fc0a6c",null);e["default"]=d.exports}}]);