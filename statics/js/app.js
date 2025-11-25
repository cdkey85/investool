$(document).ready(function () {
  // 初始化 materialize
  M.AutoInit();

  // 导航栏激活
  var currentNav = $(location).attr("pathname");
  if (currentNav.startsWith("/fund")) {
    $("#nav-fund").addClass("active");
    $("#nav-fund").siblings().removeClass("active");
  } else if (currentNav == "/about") {
    $("#nav-about").addClass("active");
    $("#nav-about").siblings().removeClass("active");
  } else if (currentNav == "/comment") {
    $("#nav-comment").addClass("active");
    $("#nav-comment").siblings().removeClass("active");
  } else if (currentNav == "/materials") {
    $("#nav-materials").addClass("active");
    $("#nav-materials").siblings().removeClass("active");
  } else {
    $("#nav-stock").addClass("active");
    $("#nav-stock").siblings().removeClass("active");
  }

  // 筛选表单中开关显示检测表单
  $("#selector_with_checker").click(function () {
    $("#checker_options").toggleClass("hide");
  });

  var beautify_value = function (value, title) {
    if (value == null) return "";
    
    if (typeof value === 'number') {
      var yi = value / 100000000.0;
      var wan = value / 10000.0;
      if (Math.abs(yi) >= 0.1) {
        return yi.toFixed(2) + "亿";
      } else if (Math.abs(wan) >= 1) {
        return wan.toFixed(2) + "万";
      }

      if (Number.isInteger(value))
        return value.toString();
      else
        return value.toFixed(2);
    }
    
    return value.toString();
  };

  // 基本面选股请求处理
  $("#selector_submit_btn").click(function () {
    $(this).addClass("disabled");
    $("#model_header").text($(this).text() + "中，请稍候...");
    $("#load_modal").modal()[0].M_Modal.options.dismissible = false;
    $("#load_modal").modal("open");
    var url = $("#selector_submit_btn").attr("actionurl");
    $.ajax({
      url: url,
      type: "post",
      data: $("#selector_form").serialize(),
      success: function (data) {
        if (data.Error) {
          $("#err_msg").text(data.Error);
          $("#error_modal").modal("open");
          $("#selector_submit_btn").removeClass("disabled");
          $("#load_modal").modal("close");
          return;
        }
        
        // 初始化列信息
        if (data.Columns && data.Columns.length > 0) {
          // 清空现有的下拉选项
          $(".dropdown-content").empty();
          
          // 先销毁已存在的 dropdown 实例（如果存在）
          $(".dropdown-trigger").dropdown('destroy');
          
          // 重新生成下拉选项
          $.each(data.Columns, function (i, column) {
            var checkboxId = "sf_" + (i + 1);
            var classId = "st_" + (i + 1);
            var columnIndex = i + 1;
            
            // 添加下拉选项
            $(".dropdown-content").append(
              '<li><label><input id="' + checkboxId + '" type="checkbox" data-column-index="' + columnIndex + '" /><span>' + column.title + '</span></label></li>'
            );
          });
          
          // 重新初始化 dropdown 组件
          $(".dropdown-trigger").dropdown({
            constrainWidth: true,
            closeOnClick: false,
          });
          
          // 在下拉框组件初始化后再绑定事件和设置默认选中状态，确保元素已经正确渲染
          setTimeout(function() {
            $.each(data.Columns, function (i, column) {
              var checkboxId = "sf_" + (i + 1);
              var classId = "st_" + (i + 1);
              var columnIndex = i + 1;
              
              // 绑定事件 - 使用事件委托确保动态元素能正确绑定事件
              $("#" + checkboxId).off('change').on('change', function () {
                var columnIndex = $(this).data("column-index");
                // 根据复选框状态切换对应列的显示/隐藏
                if (this.checked) {
                  $("." + classId).removeClass("hide");
                  $("#selector_result thead th:nth-child(" + columnIndex + ")").removeClass("hide");
                  $("#selector_result tbody tr").each(function() {
                    $(this).find("td:nth-child(" + columnIndex + ")").removeClass("hide");
                  });
                } else {
                  $("." + classId).addClass("hide");
                  $("#selector_result thead th:nth-child(" + columnIndex + ")").addClass("hide");
                  $("#selector_result tbody tr").each(function() {
                    $(this).find("td:nth-child(" + columnIndex + ")").addClass("hide");
                  });
                }
              });
              
              // 设置默认选中状态
              if (column.default_show) {
                // 使用多种方法确保复选框被正确选中
                $("#" + checkboxId)[0].checked = true;
                $("#" + checkboxId).prop("checked", true);
                $("#" + checkboxId).attr("checked", "checked");
                
                // 强制更新Materialize CSS下拉框的视觉状态
                setTimeout(function() {
                  $("#" + checkboxId).trigger('change');
                }, 50);
              }
            });
          }, 100);
        }
        
        if (data.Stocks.length == 0) {
          $(".dropdown-structure").addClass("hide");
          $("#selector_result #result_table").html(
            '<div class="row"><p class="center flow-text">无法找到符合条件的股票</p></div>'
          );
        } else {
          // 清空现有的表头
          $("#selector_result thead tr").empty();
          
          // 重新生成表头
          if (data.Columns && data.Columns.length > 0) {
            $.each(data.Columns, function (i, column) {
              var classId = "st_" + (i + 1);
              // 根据default_show决定是否添加hide类
              var thClass = column.default_show ? "" : "hide " + classId;
              $("#selector_result thead tr").append('<th class="' + thClass + '">' + column.title + '</th>');
            });
          }
          
          // 清空现有的表格内容
          $("#selector_result tbody").empty();

          $.each(data.Stocks, function (i, stock) {
          // 生成表格内容
          var row = "<tr>";
            // 根据列信息生成表格行
            if (data.Columns && data.Columns.length > 0) {
              $.each(data.Columns, function (j, column) {
                var classId = "st_" + (j + 1);
                // 根据default_show决定是否添加hide类
                var tdClass = column.default_show ? "" : "hide " + classId;
                var value = stock[column.key];
                
                // 特殊处理某些字段
                if (column.key === "SECURITY_CODE") {
                  row += '<td class="' + tdClass + '"><span class="copybtn waves-effect waves-red" data-clipboard-text="' + stock.SECURITY_CODE + '">' + stock.SECURITY_CODE + '</span></td>';
                } else if (column.key === "SECURITY_NAME_ABBR") {
                  row += '<td class="' + tdClass + '"><a target="_blank" href="http://basic.10jqka.com.cn/' + stock.SECURITY_CODE + '"/>' + stock.SECURITY_NAME_ABBR + "</a></td>";
                } else {
                  row += '<td class="' + tdClass + '">' + beautify_value(value, column.title) + "</td>";
                }
              });
            }
            
            row += "</tr>";
            $("#selector_result tbody").append(row);
          });
        }
        $("title").text(data.PageTitle);
        $("#stock_forms").remove();
        $("#selector_result").removeClass("hide");
        $("html, body").animate({ scrollTop: 0 }, 0);
        $("#load_modal").modal("close");
      },
    });
  });

  // 个股检测请求处理
  $("#checker_submit_btn").click(function () {
    if ($("#checker_keyword").val() == "") {
      $("#err_msg").text("请填写股票代码或简称");
      $("#error_modal").modal("open");
      return;
    }
    $(this).addClass("disabled");
    $("#model_header").text($(this).text() + "中，请稍候...");
    $("#load_modal").modal()[0].M_Modal.options.dismissible = false;
    $("#load_modal").modal("open");
    var url = $("#checker_submit_btn").attr("actionurl");
    $.ajax({
      url: url,
      type: "post",
      data: $("#checker_form").serialize(),
      success: function (data) {
        if (data.Error) {
          $("#err_msg").text(data.Error);
          $("#error_modal").modal("open");
          $("#checker_submit_btn").removeClass("disabled");
          $("#load_modal").modal("close");
          return;
        }
        $("title").text(data.PageTitle);
        $("#stock_forms").remove();
        $("#checker_results").removeClass("hide");
        if (data.Results.length == 0) {
          $("#checker_results h4").text("暂不支持对该股进行检测");
        } else {
          $.each(data.Results, function (i, result) {
            var cm = data.StockNames[i].split("-")[1].split(".");
            $("#checker_results").append(
              '<br/><div class="divider"></div><br/><div id="checker_result_' +
                i +
                '"><div class="row"><a target="_blank" href="http://quote.eastmoney.com/' +
                cm[1] +
                cm[0] +
                '.html">' +
                data.StockNames[i] +
                "</a><br/>当前检测财报数据来源:" +
                data.FinaReportNames[i] +
                "<br/>最新财报预约发布日期:" +
                data.FinaAppointPublishDates[i] +
                "</div>" +
                '<table class="centered striped">' +
                '<thead><tr><th width="30%">指标</th><th width="40%">描述</th><th width="30%">结果</th></tr></thead>' +
                "<tbody></tbody>" +
                "</table>" +
                "</div>"
            );
            $.each(result, function (k, v) {
              okdesc = "❌";
              if (v.ok == "true") {
                okdesc = "✅";
              }
              $(`#checker_result_${i} tbody`).append(
                "<tr><td>" +
                  k +
                  "</td><td>" +
                  v.desc +
                  "</td><td>" +
                  okdesc +
                  "</td></tr>"
              );
            });
            $(`#checker_result_${i} tbody`).append(
              "<tr><td>主力资金净流入</td><td>" +
                data.MainMoneyNetInflows[i] +
                "</td><td>--</td></tr>"
            );
            $(`#checker_result_${i}`).append(
              '<div class="row">' +
                '<br><h5 class="center">年报数据趋势概览</h5>' +
                '<div class="col s12">' +
                '<div id="line-chart-' + i + '" style="width:100%;height:400px;"></div>' +
                "</div>" +
                "</div>"
            );
            var lineChart = echarts.init(document.querySelector(`#line-chart-${i}`));
            var option = {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'cross' }
                },
                legend: {
                    data: data.Lines[i].legends,
                    top: 'bottom'
                },
                toolbox: {
                    show: true,
                    feature: {
                        dataView: {},
                        magicType: { type: ['bar'] },
                        restore: {},
                    }
                },
                xAxis: {
                    data: data.Lines[i].xAxis
                },
                yAxis: {
                    type: 'value'
                },
                series: [
                    {
                        name: "ROE",
                        type: 'line',
                        data: data.Lines[i].data[0]
                    },
                    {
                        name: "EPS",
                        type: 'line',
                        data: data.Lines[i].data[1]
                    },
                    {
                        name: "ROA",
                        type: 'line',
                        data: data.Lines[i].data[2]
                    },
                    {
                        name: "毛利率",
                        type: 'line',
                        data: data.Lines[i].data[3]
                    },
                    {
                        name: "净利率",
                        type: 'line',
                        data: data.Lines[i].data[4]
                    },
                    {
                        name: "营收",
                        type: 'line',
                        data: data.Lines[i].data[5]
                    },
                    {
                        name: "毛利润",
                        type: 'line',
                        data: data.Lines[i].data[6]
                    },
                    {
                        name: "净利润",
                        type: 'line',
                        data: data.Lines[i].data[7]
                    },
                ]
            };
            lineChart.setOption(option);
            window.onresize = function() { lineChart.resize(); };
          });
        }
        $("html, body").animate({ scrollTop: 0 }, 0);
        $("#load_modal").modal("close");
      },
    });
  });

  // 返回顶部按钮
  $("#to-top").click(function () {
    $("html, body").animate({ scrollTop: 0 }, 500);
  });
  // 按钮通过点击展示
  $(".fixed-action-btn").floatingActionButton({
    hoverEnabled: false,
  });

  // 导出结果csv文件
  $(".export-result-btn").click(function (e) {
    var table = document.getElementById("selector_result_table");
    if (!table) {
      M.toast({ html: "未找到表格数据" });
      return;
    }
    
    try {
      // 创建 TableExport 实例，不生成按钮
      var exporter = new TableExport(table, {
        formats: ["csv"],
        filename: "investool-exported",
        bootstrap: false,
        exportButtons: false  // 不生成导出按钮
      });
      
      // 直接获取导出数据
      var exportData = exporter.getExportData();
      var tableKeys = Object.keys(exportData);
      
      if (tableKeys.length === 0) {
        M.toast({ html: "未找到导出数据" });
        return;
      }
      
      var tableId = tableKeys[0];
      var csvData = exportData[tableId].csv;
      
      if (!csvData) {
        M.toast({ html: "CSV数据为空" });
        return;
      }
      
      // 调用导出方法
      exporter.export2file(
        csvData.data,
        csvData.mimeType,
        csvData.filename,
        csvData.fileExtension,
        csvData.merges || [],
        csvData.RTL || false,
        csvData.sheetname || "Sheet1"
      );
      
      M.toast({ html: "导出成功！" });
      
    } catch (err) {
      console.error("导出失败:", err);
      M.toast({ html: "导出失败: " + err.message });
    }
  });

  // 下拉框设置
  $(".dropdown-trigger").dropdown({
    constrainWidth: true,
    closeOnClick: false,
  });
  $(".dropdown-content>li>a").css("color", "#000000");
  $(".dropdown-content>li>a").css("font-size", "11px");
  $(".dropdown-content>li>a").css("font-weight", "normal");

  // 点击复制
  var clipboard = new ClipboardJS(".copybtn");
  clipboard.on("success", function (e) {
    M.toast({ html: "已复制代码至剪贴板" });
  });

  // 基金字段
  for (let i = 1; i <= 23; i++) {
    $(`#f${i}`).change(function () {
      $(`.t${i}`).toggleClass("hide");
      if (this.checked) {
        localStorage.setItem(`t${i}`, 1);
      } else {
        localStorage.removeItem(`t${i}`);
      }
    });
    if (localStorage[`t${i}`] == 1) {
      $(`.t${i}`).removeClass("hide");
      $(`#f${i}`).attr("checked", "true");
    }
  }

  // 设置排序图标
  $(".sortable").click(function () {
    var s = $(this).find("a").attr("sort");
    localStorage.setItem("fund_sort", s);
  });
  var fund_sort = localStorage["fund_sort"];
  if (fund_sort === null) {
    fund_sort = "0";
  }
  $(`.sortable a[sort='${fund_sort}'] i`).removeClass("hide");

  // 基金检测表单中开关显示检测持仓股票
  $("#check_stocks").click(function () {
    $("#checker_options").toggleClass("hide");
  });

  // 基金检测提交
  $("#check_fund_submit_btn").click(function () {
    if ($("#fundcode").val() == "") {
      $("#err_msg").text("请填写基金代码");
      $("#error_modal").modal("open");
      return;
    }
    $(this).addClass("disabled");
    $("#model_header").text($(this).text() + "中，请稍候...");
    $("#load_modal").modal()[0].M_Modal.options.dismissible = false;
    $("#load_modal").modal("open");
    var url = $("#check_fund_submit_btn").attr("actionurl");
    $.ajax({
      url: url,
      type: "post",
      data: $("#fundcheck_form").serialize(),
      success: function (data) {
        if (data.Error) {
          $("#err_msg").text(data.Error);
          $("#error_modal").modal("open");
          $("#check_fund_submit_btn").removeClass("disabled");
          $("#load_modal").modal("close");
          return;
        }

        $("title").text(data.PageTitle);
        $("#index_content").remove();
        $("#fund_check_results").removeClass("hide");

        $.each(data.Funds, function (code, fund) {
          var year_1_rank_ratio = "❌";
          if (
            fund.performance.year_1_rank_ratio < data.Param.year_1_rank_ratio
          ) {
            year_1_rank_ratio = "✅";
          }
          var year_2_rank_ratio = "❌";
          if (
            fund.performance.year_2_rank_ratio <
            data.Param.this_year_235_rank_ratio
          ) {
            year_2_rank_ratio = "✅";
          }
          var year_3_rank_ratio = "❌";
          if (
            fund.performance.year_3_rank_ratio <
            data.Param.this_year_235_rank_ratio
          ) {
            year_3_rank_ratio = "✅";
          }
          var year_5_rank_ratio = "❌";
          if (
            fund.performance.year_5_rank_ratio <
            data.Param.this_year_235_rank_ratio
          ) {
            year_5_rank_ratio = "✅";
          }
          var this_year_rank_ratio = "❌";
          if (
            fund.performance.this_year_rank_ratio <
            data.Param.this_year_235_rank_ratio
          ) {
            this_year_rank_ratio = "✅";
          }
          var month_6_rank_ratio = "❌";
          if (
            fund.performance.month_6_rank_ratio < data.Param.month_6_rank_ratio
          ) {
            month_6_rank_ratio = "✅";
          }
          var month_3_rank_ratio = "❌";
          if (
            fund.performance.month_3_rank_ratio < data.Param.month_3_rank_ratio
          ) {
            month_3_rank_ratio = "✅";
          }
          var min_scale = "❌";
          if (fund.net_assets_scale / 100000000.0 >= data.Param.min_scale) {
            min_scale = "✅";
          }
          var max_scale = "❌";
          if (fund.net_assets_scale / 100000000.0 <= data.Param.max_scale) {
            max_scale = "✅";
          }
          var manager = "❌";
          if (
            fund.manager.manage_days / 365.0 >=
            data.Param.min_manager_years
          ) {
            manager = "✅";
          }
          var stddev_avg135 = "❌";
          if (fund.stddev.avg_135 <= data.Param.max_135_avg_stddev) {
            stddev_avg135 = "✅";
          }
          var sharp_avg135 = "❌";
          if (fund.sharp.avg_135 >= data.Param.min_135_avg_sharp) {
            sharp_avg135 = "✅";
          }
          var maxretr_avg135 = "❌";
          if (fund.max_retracement.avg_135 <= data.Param.max_135_avg_retr) {
            maxretr_avg135 = "✅";
          }
          $("#fund_check_results").append(
            '<div class="row" id="' +
              fund.code +
              '"><h4 class="center"><a target="_blank" href="http://fund.eastmoney.com/' +
              fund.code +
              '.html">' +
              fund.name +
              "(" +
              fund.code +
              ')</a>检测结果</h4><p class="tiny center">以下所有数据与信息仅供参考，不构成投资建议</p><div class="divider"></div><table class="centered striped"><thead><tr><th width="30%">指标</th><th width="40%">描述</th><th width="30%">结果</th></tr></thead><tbody><tr><td>近1年绩效排名前' +
              data.Param.year_1_rank_ratio +
              "%</td><td>近1年绩效排名前" +
              fund.performance.year_1_rank_ratio.toFixed(2) +
              "%</td><td>" +
              year_1_rank_ratio +
              "</td></tr><tr><td>近2,3,5年及今年来绩效排名前" +
              data.Param.this_year_235_rank_ratio +
              "%</td><td>近2年绩效排名前" +
              fund.performance.year_2_rank_ratio.toFixed(2) +
              "%</td><td>" +
              year_2_rank_ratio +
              "</td></tr><tr><td>近2,3,5年及今年来绩效排名前" +
              data.Param.this_year_235_rank_ratio +
              "%</td><td>近3年绩效排名前" +
              fund.performance.year_3_rank_ratio.toFixed(2) +
              "%</td><td>" +
              year_3_rank_ratio +
              "</td></tr><tr><td>近2,3,5年及今年来绩效排名前" +
              data.Param.this_year_235_rank_ratio +
              "%</td><td>近5年绩效排名前" +
              fund.performance.year_5_rank_ratio.toFixed(2) +
              "%</td><td>" +
              year_5_rank_ratio +
              "</td></tr><tr><td>近2,3,5年及今年来绩效排名前" +
              data.Param.this_year_235_rank_ratio +
              "%</td><td>今年来绩效排名前" +
              fund.performance.this_year_rank_ratio.toFixed(2) +
              "%</td><td>" +
              this_year_rank_ratio +
              "</td></tr><tr><td>近6个月绩效排名前" +
              data.Param.month_6_rank_ratio +
              "%</td><td>近6个月绩效排名前" +
              fund.performance.month_6_rank_ratio.toFixed(2) +
              "%</td><td>" +
              month_6_rank_ratio +
              "</td></tr><tr><td>近3个月绩效排名前" +
              data.Param.month_3_rank_ratio +
              "%</td><td>近3个月绩效排名前" +
              fund.performance.month_3_rank_ratio.toFixed(2) +
              "%</td><td>" +
              month_3_rank_ratio +
              "</td></tr><tr><td>基金规模最低" +
              data.Param.min_scale +
              "亿</td><td>基金规模" +
              (fund.net_assets_scale / 100000000.0).toFixed(2) +
              "亿</td><td>" +
              min_scale +
              "</td></tr><tr><td>基金规模最高" +
              data.Param.max_scale +
              "亿</td><td>基金规模" +
              (fund.net_assets_scale / 100000000.0).toFixed(2) +
              "亿</td><td>" +
              max_scale +
              "</td></tr><tr><td>基金经理管理该基金不低于" +
              data.Param.min_manager_years +
              '年</td><td>基金经理:<a href="https://appunit.1234567.com.cn/fundmanager/manager.html?managerid=' +
              fund.manager.id +
              '" target="_blank">' +
              fund.manager.name +
              "</a><br/>管理该基金:" +
              (fund.manager.manage_days / 365.0).toFixed(2) +
              "年<br/>任职回报:" +
              fund.manager.manage_repay.toFixed(2) +
              "%</td><td>" +
              manager +
              "</td></tr><tr><td>近1,3,5年波动率平均值不高于" +
              data.Param.max_135_avg_stddev.toFixed(2) +
              "%</td><td>近1,3,5年波动率平均值:" +
              fund.stddev.avg_135.toFixed(2) +
              "%<br/>近1年波动率:" +
              fund.stddev.year_1.toFixed(2) +
              "%<br/>近3年波动率:" +
              fund.stddev.year_3.toFixed(2) +
              "%<br/>近5年波动率:" +
              fund.stddev.year_5.toFixed(2) +
              "%</td><td>" +
              stddev_avg135 +
              "</td></tr><tr><td>近1,3,5年夏普比率平均值不低于" +
              data.Param.min_135_avg_sharp.toFixed(2) +
              "%</td><td>近1,3,5年夏普比率平均值:" +
              fund.sharp.avg_135.toFixed(2) +
              "%<br/>近1年夏普比率:" +
              fund.sharp.year_1.toFixed(2) +
              "%<br/>近3年夏普比率:" +
              fund.sharp.year_3.toFixed(2) +
              "%<br/>近5年夏普比率:" +
              fund.sharp.year_5.toFixed(2) +
              "%</td><td>" +
              sharp_avg135 +
              "</td></tr><tr><td>近1,3,5年最大回撤率平均值不高于" +
              data.Param.max_135_avg_stddev.toFixed(2) +
              "%</td><td>近1,3,5年最大回撤率平均值:" +
              fund.max_retracement.avg_135.toFixed(2) +
              "%<br/>近1年最大回撤率:" +
              fund.max_retracement.year_1.toFixed(2) +
              "%<br/>近3年最大回撤率:" +
              fund.max_retracement.year_3.toFixed(2) +
              "%<br/>近5年最大回撤率:" +
              fund.max_retracement.year_5.toFixed(2) +
              "%</td><td>" +
              maxretr_avg135 +
              "</td></tr></tbody></table>" +
              "</div>"
          );
          if (data.StockCheckResults) {
            var stockCheckResult = data.StockCheckResults[fund.code];
            $(`#${fund.code}`).append(
              '<br/><h5 class="center">持仓股票检测结果</h5>'
            );
            $.each(stockCheckResult.check_results, function (i, result) {
              var cm = stockCheckResult.names[i].split("-")[1].split(".");
              var index = i + 1;
              $(`#${fund.code}`).append(
                '<div id="checker_result_' +
                  i +
                  '"><div class="row"><div class="divider"></div><div class="col s12 m12 l6">' +
                  index +
                  '. <a target="_blank" href="http://quote.eastmoney.com/' +
                  cm[1] +
                  cm[0] +
                  '.html">' +
                  stockCheckResult.names[i] +
                  "</a> | 持仓占比:" +
                  fund.stocks[i].hold_ratio +
                  "%<br/>所属行业:" +
                  fund.stocks[i].industry +
                  " | 最新调仓:" +
                  fund.stocks[i].adjust_ratio +
                  "%" +
                  "<br/>当前检测财报数据来源:" +
                  stockCheckResult.fina_report_names[i] +
                  "<br/>最新财报预约发布日期:" +
                  stockCheckResult.fina_appoint_publish_dates[i] +
                  "</div>" +
                  '<table class="centered striped">' +
                  '<thead><tr><th width="30%">指标</th><th width="40%">描述</th><th width="30%">结果</th></tr></thead>' +
                  "<tbody></tbody>" +
                  "</table>" +
                  "</div>"
              );
              $.each(result, function (k, v) {
                okdesc = "❌";
                if (v.ok == "true") {
                  okdesc = "✅";
                }
                $(`#${fund.code} #checker_result_${i} tbody`).append(
                  "<tr><td>" +
                    k +
                    "</td><td>" +
                    v.desc +
                    "</td><td>" +
                    okdesc +
                    "</td></tr>"
                );
              });
            });
          }
        });
        $("html, body").animate({ scrollTop: 0 }, 0);
        $("#load_modal").modal("close");
      },
    });
  });
});
