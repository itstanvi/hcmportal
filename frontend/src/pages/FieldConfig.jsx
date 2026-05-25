import { useEffect, useState } from 'react';
import api, { extractError } from '../api/client';

const FIELD_TYPES = ['text', 'number', 'date', 'dropdown', 'boolean', 'email', 'phone'];

const emptyForm = {
  field_name: '',
  field_key: '',
  field_type: 'text',
  required: false,
  active: true,
  options: '',
  display_order: 0,
};

export default function FieldConfig() {
  const [fields, setFields] = useState([]);
  const [form, setForm] = useState(emptyForm);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');

  const load = async () => {
    setLoading(true);
    try {
      const res = await api.get('/employee-fields');
      setFields(res.data || []);
    } catch (err) {
      setError(extractError(err));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const parseOptions = (s) =>
    s.split(',').map((x) => x.trim()).filter(Boolean);

  const onCreate = async (e) => {
    e.preventDefault();
    setError('');
    setInfo('');
    setSaving(true);
    try {
      const payload = {
        field_name: form.field_name,
        field_key: form.field_key,
        field_type: form.field_type,
        required: !!form.required,
        active: !!form.active,
        display_order: Number(form.display_order) || 0,
      };
      if (form.field_type === 'dropdown') {
        payload.options = parseOptions(form.options);
      }
      await api.post('/employee-fields', payload);
      setInfo('Field created.');
      setForm(emptyForm);
      load();
    } catch (err) {
      setError(extractError(err));
    } finally {
      setSaving(false);
    }
  };

  const updateField = async (field, patch) => {
    setError('');
    setInfo('');
    try {
      await api.put(`/employee-fields/${field.id}`, patch);
      load();
    } catch (err) {
      setError(extractError(err));
    }
  };

  const onEditOptions = async (field) => {
    let current = [];
    if (Array.isArray(field.options)) current = field.options;
    else if (typeof field.options === 'string') {
      try { current = JSON.parse(field.options); } catch { /* noop */ }
    }
    const next = window.prompt('Comma separated options', current.join(', '));
    if (next === null) return;
    const opts = parseOptions(next);
    await updateField(field, { options: opts });
  };

  const onRename = async (field) => {
    const next = window.prompt('Field display name', field.field_name);
    if (next === null || !next.trim()) return;
    await updateField(field, { field_name: next.trim() });
  };

  const onChangeOrder = async (field) => {
    const next = window.prompt('Display order (number)', String(field.display_order));
    if (next === null) return;
    const n = Number(next);
    if (Number.isNaN(n)) return;
    await updateField(field, { display_order: n });
  };

  return (
    <>
      <div className="page-header">
        <h2>Employee Field Configuration</h2>
      </div>

      {error && <div className="alert error">{error}</div>}
      {info && <div className="alert success">{info}</div>}

      <div className="card">
        <div className="section-title" style={{ marginTop: 0 }}>Add a new field</div>
        <form onSubmit={onCreate}>
          <div className="form-grid">
            <div>
              <label>Field Name<span className="required-mark">*</span></label>
              <input
                value={form.field_name}
                onChange={(e) => setForm({ ...form, field_name: e.target.value })}
                required
              />
            </div>
            <div>
              <label>Field Key<span className="required-mark">*</span></label>
              <input
                value={form.field_key}
                onChange={(e) => setForm({ ...form, field_key: e.target.value })}
                placeholder="e.g. work_mode (lowercase, snake_case)"
                required
              />
            </div>
            <div>
              <label>Field Type</label>
              <select
                value={form.field_type}
                onChange={(e) => setForm({ ...form, field_type: e.target.value })}
              >
                {FIELD_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
              </select>
            </div>
            <div>
              <label>Display Order</label>
              <input
                type="number"
                value={form.display_order}
                onChange={(e) => setForm({ ...form, display_order: e.target.value })}
              />
            </div>
            {form.field_type === 'dropdown' && (
              <div className="full">
                <label>Dropdown Options (comma-separated)</label>
                <input
                  value={form.options}
                  onChange={(e) => setForm({ ...form, options: e.target.value })}
                  placeholder="Office, Hybrid, Remote"
                />
              </div>
            )}
            <div className="full flex gap-16">
              <label style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 0 }}>
                <input
                  type="checkbox"
                  style={{ width: 'auto' }}
                  checked={form.required}
                  onChange={(e) => setForm({ ...form, required: e.target.checked })}
                />
                Required
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 0 }}>
                <input
                  type="checkbox"
                  style={{ width: 'auto' }}
                  checked={form.active}
                  onChange={(e) => setForm({ ...form, active: e.target.checked })}
                />
                Active
              </label>
            </div>
          </div>
          <div className="mt-16">
            <button type="submit" className="primary" disabled={saving}>
              {saving ? 'Saving…' : 'Add Field'}
            </button>
          </div>
        </form>
      </div>

      <div className="section-title">Existing Fields</div>
      {loading ? (
        <div className="empty">Loading…</div>
      ) : fields.length === 0 ? (
        <div className="empty">No custom fields configured yet.</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Order</th>
              <th>Name</th>
              <th>Key</th>
              <th>Type</th>
              <th>Required</th>
              <th>Active</th>
              <th>Options</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {fields.map((f) => (
              <tr key={f.id} style={{ cursor: 'default' }}>
                <td>{f.display_order}</td>
                <td>{f.field_name}</td>
                <td><code>{f.field_key}</code></td>
                <td><span className="badge">{f.field_type}</span></td>
                <td>{f.required ? 'Yes' : 'No'}</td>
                <td>
                  <span className={`badge ${f.active ? 'success' : 'muted'}`}>
                    {f.active ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td>
                  {f.field_type === 'dropdown'
                    ? (Array.isArray(f.options) ? f.options.join(', ') : (f.options || '—'))
                    : '—'}
                </td>
                <td>
                  <div className="flex gap-8" onClick={(e) => e.stopPropagation()}>
                    <button onClick={() => onRename(f)}>Rename</button>
                    <button onClick={() => onChangeOrder(f)}>Order</button>
                    {f.field_type === 'dropdown' && (
                      <button onClick={() => onEditOptions(f)}>Options</button>
                    )}
                    <button onClick={() => updateField(f, { required: !f.required })}>
                      {f.required ? 'Make optional' : 'Make required'}
                    </button>
                    <button onClick={() => updateField(f, { active: !f.active })}>
                      {f.active ? 'Deactivate' : 'Activate'}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}
