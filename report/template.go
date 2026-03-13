package report

const reportTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>AWR vs WDR Performance Comparison Report</title>
<style>
  :root {
    --oracle-color: #c74634;
    --gauss-color: #2b7ae8;
    --bg-primary: #f8f9fa;
    --bg-white: #ffffff;
    --border-color: #e2e8f0;
    --text-primary: #1a202c;
    --text-secondary: #64748b;
    --green: #16a34a;
    --red: #dc2626;
    --yellow: #f59e0b;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "PingFang SC", "Microsoft YaHei", sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    line-height: 1.6;
  }
  .header {
    background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
    color: white;
    padding: 32px 40px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .header h1 { font-size: 22px; font-weight: 600; letter-spacing: 0.5px; }
  .header .subtitle { font-size: 13px; color: #94a3b8; margin-top: 4px; }
  .header .generated { font-size: 12px; color: #94a3b8; text-align: right; }
  .legend {
    display: flex; gap: 24px; padding: 16px 40px;
    background: var(--bg-white); border-bottom: 1px solid var(--border-color);
    font-size: 13px; color: var(--text-secondary);
  }
  .legend-item { display: flex; align-items: center; gap: 6px; }
  .legend-dot { width: 10px; height: 10px; border-radius: 50%; }
  .legend-dot.oracle { background: var(--oracle-color); }
  .legend-dot.gauss { background: var(--gauss-color); }
  .legend-dot.better { background: var(--green); }
  .legend-dot.worse { background: var(--red); }
  .container { max-width: 1400px; margin: 0 auto; padding: 24px 40px 60px; }
  .section { margin-bottom: 32px; }
  .section-title {
    font-size: 16px; font-weight: 600; color: var(--text-primary);
    margin-bottom: 16px; padding-left: 12px; border-left: 3px solid #334155;
  }
  .instance-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
  .instance-card { background: var(--bg-white); border-radius: 8px; border: 1px solid var(--border-color); overflow: hidden; }
  .instance-card-header {
    padding: 12px 20px; font-size: 14px; font-weight: 600; color: white;
    display: flex; align-items: center; gap: 8px;
  }
  .instance-card-header.oracle { background: var(--oracle-color); }
  .instance-card-header.gauss { background: var(--gauss-color); }
  .instance-card-body { padding: 0; }
  .instance-card-body table { width: 100%; border-collapse: collapse; }
  .instance-card-body td { padding: 10px 20px; font-size: 13px; border-bottom: 1px solid #f1f5f9; }
  .instance-card-body tr:last-child td { border-bottom: none; }
  .instance-card-body td:first-child { color: var(--text-secondary); width: 140px; font-weight: 500; }
  .instance-card-body td:last-child { font-weight: 600; }
  .comp-table-wrapper { background: var(--bg-white); border-radius: 8px; border: 1px solid var(--border-color); overflow-x: auto; }
  .comp-table { width: 100%; border-collapse: collapse; font-size: 13px; }
  .comp-table thead th {
    padding: 12px 16px; text-align: left; font-weight: 600; font-size: 12px;
    text-transform: uppercase; letter-spacing: 0.5px; color: var(--text-secondary);
    background: #f8fafc; border-bottom: 2px solid var(--border-color); white-space: nowrap;
  }
  .comp-table thead th.col-oracle { background: #fef2f0; color: var(--oracle-color); }
  .comp-table thead th.col-gauss { background: #eff6ff; color: var(--gauss-color); }
  .comp-table thead th.col-diff { background: #f8fafc; text-align: center; }
  .comp-table tbody td { padding: 10px 16px; border-bottom: 1px solid #f1f5f9; vertical-align: top; }
  .comp-table tbody tr:hover { background: #f8fafc; }
  .comp-table tbody tr:last-child td { border-bottom: none; }
  .comp-table .rank { color: var(--text-secondary); font-weight: 500; text-align: center; width: 36px; }
  .comp-table .metric-name { font-weight: 600; max-width: 220px; }
  .comp-table .num {
    text-align: right; font-variant-numeric: tabular-nums;
    font-family: "SF Mono", "Fira Code", "Cascadia Code", monospace;
  }
  .diff-badge {
    display: inline-block; padding: 2px 8px; border-radius: 10px;
    font-size: 11px; font-weight: 600; text-align: center; min-width: 56px;
  }
  .diff-badge.better { background: #dcfce7; color: var(--green); }
  .diff-badge.worse { background: #fef2f2; color: var(--red); }
  .diff-badge.neutral { background: #fefce8; color: var(--yellow); }
  .diff-cell { text-align: center; }
  .comp-table .divider-left { border-left: 2px solid var(--border-color); }
  .sql-item { background: var(--bg-white); border-radius: 8px; border: 1px solid var(--border-color); margin-bottom: 16px; overflow: hidden; }
  .sql-item-header {
    display: flex; align-items: center; padding: 12px 20px;
    background: #f8fafc; border-bottom: 1px solid var(--border-color);
  }
  .sql-rank-badge {
    background: #334155; color: white; width: 28px; height: 28px; border-radius: 50%;
    display: flex; align-items: center; justify-content: center;
    font-size: 13px; font-weight: 600; flex-shrink: 0;
  }
  .sql-item-title { flex: 1; margin-left: 12px; font-size: 13px; display: flex; gap: 16px; align-items: center; }
  .sql-id-tag { font-family: "SF Mono", monospace; font-size: 12px; padding: 2px 8px; border-radius: 4px; font-weight: 600; }
  .sql-id-tag.oracle { background: #fef2f0; color: var(--oracle-color); }
  .sql-id-tag.gauss { background: #eff6ff; color: var(--gauss-color); }
  .sql-metrics-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1px; background: var(--border-color); }
  .sql-metric-cell { background: var(--bg-white); padding: 14px 20px; text-align: center; }
  .sql-metric-cell .label {
    font-size: 11px; color: var(--text-secondary); text-transform: uppercase;
    letter-spacing: 0.5px; margin-bottom: 8px;
  }
  .sql-metric-values { display: flex; justify-content: center; align-items: center; gap: 12px; }
  .sql-metric-val {
    font-size: 15px; font-weight: 700; font-variant-numeric: tabular-nums;
    font-family: "SF Mono", monospace;
  }
  .sql-metric-val.oracle { color: var(--oracle-color); }
  .sql-metric-val.gauss { color: var(--gauss-color); }
  .sql-metric-vs { font-size: 11px; color: var(--text-secondary); font-weight: 500; }
  .sql-text-area { padding: 16px 20px; border-top: 1px solid var(--border-color); }
  .sql-text-label {
    font-size: 11px; color: var(--text-secondary); text-transform: uppercase;
    letter-spacing: 0.5px; margin-bottom: 8px;
  }
  .sql-text-content {
    font-family: "SF Mono", "Fira Code", monospace; font-size: 12px; line-height: 1.7;
    background: #f8fafc; padding: 12px 16px; border-radius: 6px; color: #334155;
    white-space: pre-wrap; word-break: break-all; max-height: 120px; overflow-y: auto;
  }
  .sql-text-dual { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .sql-text-single { margin-bottom: 8px; }
  .sql-text-tag { font-size: 11px; font-weight: 600; margin-bottom: 4px; }
  .sql-text-tag.oracle { color: var(--oracle-color); }
  .sql-text-tag.gauss { color: var(--gauss-color); }
  .similarity-badge {
    font-size: 11px; padding: 2px 8px; border-radius: 10px;
    background: #f0f9ff; color: #0369a1; font-weight: 600;
  }
  .footer { text-align: center; padding: 24px; font-size: 12px; color: var(--text-secondary); border-top: 1px solid var(--border-color); }
  @media (max-width: 900px) {
    .instance-grid { grid-template-columns: 1fr; }
    .sql-metrics-grid { grid-template-columns: repeat(2, 1fr); }
    .sql-text-dual { grid-template-columns: 1fr; }
    .container { padding: 16px; }
    .header { padding: 24px 16px; flex-direction: column; gap: 8px; }
  }
</style>
</head>
<body>

<div class="header">
  <div>
    <h1>AWR vs WDR Performance Comparison Report</h1>
    <div class="subtitle">Oracle &rarr; openGauss Migration Performance Analysis</div>
  </div>
  <div class="generated">
    Generated: {{.GeneratedAt}}<br>
    Tool: awr2wdr v1.0
  </div>
</div>

<div class="legend">
  <div class="legend-item"><div class="legend-dot oracle"></div>Oracle</div>
  <div class="legend-item"><div class="legend-dot gauss"></div>openGauss</div>
  <div class="legend-item"><div class="legend-dot better"></div>openGauss 更优</div>
  <div class="legend-item"><div class="legend-dot worse"></div>openGauss 较差</div>
</div>

<div class="container">

  <!-- Instance Summary -->
  <div class="section">
    <div class="section-title">实例汇总信息</div>
    <div class="instance-grid">
      <div class="instance-card">
        <div class="instance-card-header oracle">Oracle Database</div>
        <div class="instance-card-body">
          <table>
            <tr><td>数据库名</td><td>{{.Oracle.Instance.DBName}}</td></tr>
            <tr><td>实例名</td><td>{{.Oracle.Instance.InstanceName}}</td></tr>
            <tr><td>版本</td><td>{{.Oracle.Instance.Version}}</td></tr>
            <tr><td>主机名</td><td>{{.Oracle.Instance.HostName}}</td></tr>
            <tr><td>快照范围</td><td>{{.Oracle.Instance.SnapStart}} ~ {{.Oracle.Instance.SnapEnd}}</td></tr>
            <tr><td>DB Time</td><td>{{formatFloat .Oracle.Instance.DBTime}} min</td></tr>
            <tr><td>Elapsed Time</td><td>{{formatFloat .Oracle.Instance.ElapsedTime}} min</td></tr>
            <tr><td>DB CPU</td><td>{{formatFloat .Oracle.Instance.DBCPU}} min</td></tr>
            <tr><td>平均 QPS</td><td>{{formatFloat .Oracle.Instance.QPS}}</td></tr>
            <tr><td>平均 TPS</td><td>{{formatFloat .Oracle.Instance.TPS}}</td></tr>
            <tr><td>平均 DB Time/s</td><td>{{formatFloat .Oracle.Instance.AvgDBTime}} s</td></tr>
            <tr><td>平均 CPU Time/s</td><td>{{formatFloat .Oracle.Instance.AvgCPUTime}} s</td></tr>
          </table>
        </div>
      </div>
      <div class="instance-card">
        <div class="instance-card-header gauss">openGauss</div>
        <div class="instance-card-body">
          <table>
            <tr><td>数据库名</td><td>{{.Gauss.Instance.DBName}}</td></tr>
            <tr><td>实例名</td><td>{{.Gauss.Instance.InstanceName}}</td></tr>
            <tr><td>版本</td><td>{{.Gauss.Instance.Version}}</td></tr>
            <tr><td>主机名</td><td>{{.Gauss.Instance.HostName}}</td></tr>
            <tr><td>快照范围</td><td>{{.Gauss.Instance.SnapStart}} ~ {{.Gauss.Instance.SnapEnd}}</td></tr>
            <tr><td>DB Time</td><td>{{formatFloat .Gauss.Instance.DBTime}} min</td></tr>
            <tr><td>Elapsed Time</td><td>{{formatFloat .Gauss.Instance.ElapsedTime}} min</td></tr>
            <tr><td>DB CPU</td><td>{{formatFloat .Gauss.Instance.DBCPU}} min</td></tr>
            <tr><td>平均 QPS</td><td>{{formatFloat .Gauss.Instance.QPS}}</td></tr>
            <tr><td>平均 TPS</td><td>{{formatFloat .Gauss.Instance.TPS}}</td></tr>
            <tr><td>平均 DB Time/s</td><td>{{formatFloat .Gauss.Instance.AvgDBTime}} s</td></tr>
            <tr><td>平均 CPU Time/s</td><td>{{formatFloat .Gauss.Instance.AvgCPUTime}} s</td></tr>
          </table>
        </div>
      </div>
    </div>
  </div>

  <!-- Wait Events -->
  <div class="section">
    <div class="section-title">Top 等待事件对比</div>
    <div class="comp-table-wrapper">
      <table class="comp-table">
        <thead>
          <tr>
            <th style="width:36px">#</th>
            <th class="col-oracle">等待事件<br><small>Oracle</small></th>
            <th class="col-oracle">等待类别</th>
            <th class="col-oracle num">等待次数</th>
            <th class="col-oracle num">总等待时间(s)</th>
            <th class="col-oracle num">占DB Time%</th>
            <th class="col-gauss divider-left">等待事件<br><small>openGauss</small></th>
            <th class="col-gauss">等待类别</th>
            <th class="col-gauss num">等待次数</th>
            <th class="col-gauss num">总等待时间(s)</th>
            <th class="col-gauss num">占DB Time%</th>
            <th class="col-diff divider-left">时间差异</th>
          </tr>
        </thead>
        <tbody>
          {{range .WaitPairs}}
          <tr>
            <td class="rank">{{.Rank}}</td>
            <td class="metric-name">{{.OracleEvent.EventName}}</td>
            <td>{{.OracleEvent.WaitClass}}</td>
            <td class="num">{{formatInt .OracleEvent.Waits}}</td>
            <td class="num">{{formatFloat .OracleEvent.TotalTime}}</td>
            <td class="num">{{formatFloat .OracleEvent.PctDBTime}}%</td>
            <td class="metric-name divider-left">{{.GaussEvent.EventName}}</td>
            <td>{{.GaussEvent.WaitClass}}</td>
            <td class="num">{{formatInt .GaussEvent.Waits}}</td>
            <td class="num">{{formatFloat .GaussEvent.TotalTime}}</td>
            <td class="num">{{formatFloat .GaussEvent.PctDBTime}}%</td>
            <td class="diff-cell divider-left"><span class="diff-badge {{diffClass .TimeDiffPct}}">{{formatDiffPct .TimeDiffPct}}</span></td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>

  <!-- Top SQL -->
  <div class="section">
    <div class="section-title">Top SQL 对比（按执行次数排名）</div>
    {{range .SQLPairs}}
    <div class="sql-item">
      <div class="sql-item-header">
        <div class="sql-rank-badge">{{.Rank}}</div>
        <div class="sql-item-title">
          <span class="sql-id-tag oracle">SQL_ID: {{.OracleSQL.SQLID}}</span>
          <span class="sql-id-tag gauss">Query ID: {{.GaussSQL.SQLID}}</span>
          <span class="similarity-badge">相似度: {{formatPct .Similarity}}</span>
          <span style="color:#64748b;font-size:12px;">{{truncateSQL .OracleSQL.SQLText 80}}</span>
        </div>
      </div>
      <div class="sql-metrics-grid">
        <div class="sql-metric-cell">
          <div class="label">执行次数</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatInt .OracleSQL.Executions}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatInt .GaussSQL.Executions}}</span>
          </div>
        </div>
        <div class="sql-metric-cell">
          <div class="label">平均响应时间 (ms)</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatFloat .OracleSQL.AvgElapsed}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatFloat .GaussSQL.AvgElapsed}}</span>
          </div>
          {{if gt .OracleSQL.AvgElapsed 0.0}}
          <div style="margin-top:6px;"><span class="diff-badge {{diffClass (diffPct .OracleSQL.AvgElapsed .GaussSQL.AvgElapsed)}}">{{formatDiffPct (diffPct .OracleSQL.AvgElapsed .GaussSQL.AvgElapsed)}}</span></div>
          {{end}}
        </div>
        <div class="sql-metric-cell">
          <div class="label">平均 CPU 时间 (ms)</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatFloat .OracleSQL.AvgCPUTime}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatFloat .GaussSQL.AvgCPUTime}}</span>
          </div>
          {{if gt .OracleSQL.AvgCPUTime 0.0}}
          <div style="margin-top:6px;"><span class="diff-badge {{diffClass (diffPct .OracleSQL.AvgCPUTime .GaussSQL.AvgCPUTime)}}">{{formatDiffPct (diffPct .OracleSQL.AvgCPUTime .GaussSQL.AvgCPUTime)}}</span></div>
          {{end}}
        </div>
        <div class="sql-metric-cell">
          <div class="label">平均逻辑读</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatInt .OracleSQL.AvgLogicalRead}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatInt .GaussSQL.AvgLogicalRead}}</span>
          </div>
          {{if gt .OracleSQL.AvgLogicalRead 0}}
          <div style="margin-top:6px;"><span class="diff-badge {{diffClass (diffPctInt .OracleSQL.AvgLogicalRead .GaussSQL.AvgLogicalRead)}}">{{formatDiffPct (diffPctInt .OracleSQL.AvgLogicalRead .GaussSQL.AvgLogicalRead)}}</span></div>
          {{end}}
        </div>
        <div class="sql-metric-cell">
          <div class="label">平均物理读</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatInt .OracleSQL.AvgPhysicalRead}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatInt .GaussSQL.AvgPhysicalRead}}</span>
          </div>
          {{if gt .OracleSQL.AvgPhysicalRead 0}}
          <div style="margin-top:6px;"><span class="diff-badge {{diffClass (diffPctInt .OracleSQL.AvgPhysicalRead .GaussSQL.AvgPhysicalRead)}}">{{formatDiffPct (diffPctInt .OracleSQL.AvgPhysicalRead .GaussSQL.AvgPhysicalRead)}}</span></div>
          {{end}}
        </div>
        <div class="sql-metric-cell">
          <div class="label">平均返回行数</div>
          <div class="sql-metric-values">
            <span class="sql-metric-val oracle">{{formatInt .OracleSQL.AvgRows}}</span>
            <span class="sql-metric-vs">vs</span>
            <span class="sql-metric-val gauss">{{formatInt .GaussSQL.AvgRows}}</span>
          </div>
          {{if gt .OracleSQL.AvgRows 0}}
          <div style="margin-top:6px;"><span class="diff-badge {{diffClass (diffPctInt .OracleSQL.AvgRows .GaussSQL.AvgRows)}}">{{formatDiffPct (diffPctInt .OracleSQL.AvgRows .GaussSQL.AvgRows)}}</span></div>
          {{end}}
        </div>
      </div>
      <div class="sql-text-area">
        <div class="sql-text-label">SQL 文本</div>
        <div class="sql-text-dual">
          <div>
            <div class="sql-text-tag oracle">Oracle</div>
            <div class="sql-text-content">{{.OracleSQL.SQLText}}</div>
          </div>
          <div>
            <div class="sql-text-tag gauss">openGauss</div>
            <div class="sql-text-content">{{.GaussSQL.SQLText}}</div>
          </div>
        </div>
      </div>
    </div>
    {{end}}
  </div>

</div>

<div class="footer">
  Generated by <strong>awr2wdr</strong> &mdash; Oracle AWR vs openGauss WDR Performance Comparison Tool
</div>

</body>
</html>`
