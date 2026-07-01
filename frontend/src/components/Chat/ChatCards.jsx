// Data card components for different intent types

export function DataCard({ intent, data }) {
  if (!data) return null;

  if (data.id && data.title && !Array.isArray(data)) {
    return <SingleWorkItemCard item={data} />;
  }
  if (Array.isArray(data) && data.length > 0 && data[0].id) {
    return <WorkItemListCard items={data} />;
  }
  if (data.sprint && data.total !== undefined) {
    return <SprintSummaryCard data={data} />;
  }
  if (data.distribution) {
    return <TeamWorkloadCard data={data} />;
  }
  if (data.user && data.task_count !== undefined) {
    return <UserWorkloadCard data={data} />;
  }
  if (data.score !== undefined && data.status) {
    return <SprintHealthCard data={data} />;
  }
  return null;
}

function SingleWorkItemCard({ item }) {
  const formatDate = (d) => d ? new Date(d).toLocaleDateString('en-IN', { day: 'numeric', month: 'short', year: 'numeric' }) : '—';
  const stripHtml = (html) => {
    if (!html) return null;
    return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim().slice(0, 300) || null;
  };
  const desc = stripHtml(item.description);

  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', padding: '20px', overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '12px' }}>
        <span style={{ fontSize: '11px', padding: '2px 8px', borderRadius: '6px', backgroundColor: '#10b981', color: 'white', fontWeight: 600 }}>{item.workItemType || 'Item'}</span>
        <span style={{ fontSize: '12px', color: '#737373' }}>#{item.id}</span>
      </div>
      <div style={{ fontSize: '16px', fontWeight: 600, color: '#fafafa', marginBottom: '14px', lineHeight: 1.4 }}>{item.title}</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', fontSize: '13px' }}>
        <Field label="Status" value={item.state} />
        <Field label="Assigned To" value={item.assignedTo || 'Unassigned'} />
        <Field label="Priority" value={item.priority > 0 ? `P${item.priority}` : '—'} />
        <Field label="Sprint" value={item.iterationPath ? item.iterationPath.split('\\').pop() : '—'} />
        <Field label="Created" value={formatDate(item.createdDate)} />
        <Field label="Updated" value={formatDate(item.changedDate)} />
      </div>
      {desc && (
        <div style={{ marginTop: '14px', paddingTop: '14px', borderTop: '1px solid #2e2e2e', fontSize: '12px', color: '#a3a3a3', lineHeight: 1.6 }}>
          {desc.length >= 300 ? desc + '...' : desc}
        </div>
      )}
      {item.url && (
        <a href={item.url} target="_blank" rel="noopener noreferrer" style={{ display: 'inline-block', marginTop: '12px', fontSize: '12px', color: '#34d399', textDecoration: 'none' }}>
          View in Azure DevOps →
        </a>
      )}
    </div>
  );
}

function WorkItemListCard({ items }) {
  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', overflow: 'hidden' }}>
      <div style={{ padding: '14px 20px', borderBottom: '1px solid #2e2e2e', fontSize: '13px', color: '#a3a3a3', fontWeight: 500 }}>
        {items.length} work item{items.length > 1 ? 's' : ''}
      </div>
      <div style={{ maxHeight: '320px', overflowY: 'auto' }}>
        {items.map((item, i) => (
          <div key={item.id || i} style={{ padding: '12px 20px', borderBottom: i < items.length - 1 ? '1px solid #222' : 'none', display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: '13px', fontWeight: 500, color: '#e5e5e5', marginBottom: '4px' }}>
                <span style={{ color: '#737373', marginRight: '6px' }}>#{item.id}</span>
                {item.title}
              </div>
              <div style={{ fontSize: '11px', color: '#737373', display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
                <span style={{ padding: '1px 6px', borderRadius: '4px', backgroundColor: '#262626', color: '#a3a3a3' }}>{item.state}</span>
                {item.assignedTo && <span>{item.assignedTo}</span>}
                {item.workItemType && <span>{item.workItemType}</span>}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function SprintSummaryCard({ data }) {
  const progress = data.total > 0 ? Math.round((data.done / data.total) * 100) : 0;
  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', padding: '20px' }}>
      <div style={{ fontSize: '14px', fontWeight: 600, color: '#fafafa', marginBottom: '16px' }}>📊 Sprint: {data.sprint?.split('\\').pop() || data.sprint}</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: '12px', marginBottom: '16px' }}>
        <StatBox label="Total" value={data.total} />
        <StatBox label="Done" value={data.done} color="#34d399" />
        <StatBox label="Active" value={data.active} color="#fbbf24" />
        <StatBox label="Bugs" value={data.bugs} color="#f87171" />
      </div>
      <div style={{ height: '6px', backgroundColor: '#262626', borderRadius: '3px', overflow: 'hidden' }}>
        <div style={{ height: '100%', width: `${progress}%`, backgroundColor: '#34d399', borderRadius: '3px', transition: 'width 0.3s' }} />
      </div>
      <div style={{ fontSize: '12px', color: '#737373', marginTop: '6px', textAlign: 'right' }}>{progress}% complete</div>
    </div>
  );
}

function TeamWorkloadCard({ data }) {
  const entries = Object.entries(data.distribution || {}).sort((a, b) => b[1] - a[1]);
  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', padding: '20px' }}>
      <div style={{ fontSize: '14px', fontWeight: 600, color: '#fafafa', marginBottom: '14px' }}>👥 Team Workload</div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
        {entries.map(([name, count]) => {
          const max = entries[0]?.[1] || 1;
          return (
            <div key={name} style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
              <span style={{ fontSize: '12px', color: '#d4d4d4', width: '140px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{name.split(' ')[0]}</span>
              <div style={{ flex: 1, height: '6px', backgroundColor: '#262626', borderRadius: '3px' }}>
                <div style={{ height: '100%', width: `${(count / max) * 100}%`, backgroundColor: count > 5 ? '#f87171' : '#34d399', borderRadius: '3px' }} />
              </div>
              <span style={{ fontSize: '12px', color: '#737373', width: '20px', textAlign: 'right' }}>{count}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function UserWorkloadCard({ data }) {
  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', padding: '20px' }}>
      <div style={{ fontSize: '14px', fontWeight: 600, color: '#fafafa', marginBottom: '14px' }}>📊 {data.user}</div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
        <StatBox label="Active Tasks" value={data.task_count} />
        <StatBox label="Workload" value={data.level} />
      </div>
      {data.states && (
        <div style={{ marginTop: '12px', paddingTop: '12px', borderTop: '1px solid #2e2e2e', display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
          {Object.entries(data.states).map(([state, count]) => (
            <span key={state} style={{ fontSize: '11px', padding: '3px 8px', borderRadius: '6px', backgroundColor: '#262626', color: '#a3a3a3' }}>
              {state}: {count}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function SprintHealthCard({ data }) {
  const scoreColor = data.score >= 80 ? '#34d399' : data.score >= 50 ? '#fbbf24' : '#f87171';
  return (
    <div style={{ backgroundColor: '#1a1a1a', border: '1px solid #333', borderRadius: '12px', padding: '20px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
        <span style={{ fontSize: '14px', fontWeight: 600, color: '#fafafa' }}>🏥 Sprint Health</span>
        <span style={{ fontSize: '24px', fontWeight: 700, color: scoreColor }}>{data.score}/100</span>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '10px' }}>
        <StatBox label="Blocked" value={data.blocked_count || 0} color="#f87171" />
        <StatBox label="Risks" value={data.risks_count || 0} color="#fbbf24" />
        <StatBox label="Bottlenecks" value={data.bottlenecks?.length || 0} color="#34d399" />
      </div>
    </div>
  );
}

function StatBox({ label, value, color }) {
  return (
    <div style={{ backgroundColor: '#262626', borderRadius: '8px', padding: '10px 12px', textAlign: 'center' }}>
      <div style={{ fontSize: '18px', fontWeight: 700, color: color || '#fafafa' }}>{value}</div>
      <div style={{ fontSize: '11px', color: '#737373', marginTop: '2px' }}>{label}</div>
    </div>
  );
}

function Field({ label, value }) {
  return (
    <div>
      <span style={{ fontSize: '11px', color: '#737373' }}>{label}</span>
      <div style={{ fontSize: '13px', color: '#d4d4d4', marginTop: '2px' }}>{value}</div>
    </div>
  );
}
