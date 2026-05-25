import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import api, { extractError } from '../api/client';
import DynamicField from '../components/DynamicField';

const EMPLOYMENT_TYPES = ['Full-Time', 'Part-Time', 'Contract', 'Intern'];
const STATUSES = ['ACTIVE', 'INACTIVE', 'ON_LEAVE', 'TERMINATED'];

export default function EmployeeForm() {
  const { id } = useParams();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const { register, handleSubmit, reset, formState: { errors, isSubmitting } } = useForm();
  const [fields, setFields] = useState([]);
  const [loading, setLoading] = useState(true);
  const [serverError, setServerError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    let cancel = false;
    const load = async () => {
      setLoading(true);
      try {
        const fieldsRes = await api.get('/employee-fields', { params: { active_only: true } });
        if (cancel) return;
        setFields(fieldsRes.data || []);

        if (isEdit) {
          const empRes = await api.get(`/employees/${id}`);
          if (cancel) return;
          const e = empRes.data;
          reset({
            name: e.name,
            email: e.email,
            phone: e.phone,
            employee_code: e.employee_code,
            department: e.department,
            designation: e.designation,
            date_of_joining: e.date_of_joining ? e.date_of_joining.substring(0, 10) : '',
            employment_type: e.employment_type,
            status: e.status,
            custom_fields: normalizeCustomForForm(e.custom_fields),
          });
        }
      } catch (err) {
        setServerError(extractError(err));
      } finally {
        if (!cancel) setLoading(false);
      }
    };
    load();
    return () => { cancel = true; };
  }, [id, isEdit, reset]);

  // Convert API custom_fields object into form-friendly strings (booleans stored as "true"/"false").
  const normalizeCustomForForm = (cf) => {
    const out = {};
    if (!cf) return out;
    for (const [k, v] of Object.entries(cf)) {
      if (typeof v === 'boolean') out[k] = String(v);
      else out[k] = v ?? '';
    }
    return out;
  };

  const onSubmit = async (values) => {
    setServerError('');
    setSuccess('');

    // Type-coerce custom fields per the field config.
    const customFields = {};
    for (const f of fields) {
      const raw = values.custom_fields?.[f.field_key];
      if (raw === undefined || raw === '' || raw === null) continue;
      switch (f.field_type) {
        case 'number':
          customFields[f.field_key] = Number(raw);
          break;
        case 'boolean':
          customFields[f.field_key] = raw === 'true' || raw === true;
          break;
        default:
          customFields[f.field_key] = raw;
      }
    }

    const payload = {
      name: values.name,
      email: values.email,
      phone: values.phone || '',
      employee_code: values.employee_code,
      department: values.department || '',
      designation: values.designation || '',
      date_of_joining: values.date_of_joining ? new Date(values.date_of_joining).toISOString() : null,
      employment_type: values.employment_type || '',
      status: values.status || 'ACTIVE',
      custom_fields: customFields,
    };

    try {
      if (isEdit) {
        await api.put(`/employees/${id}`, payload);
        setSuccess('Employee updated successfully.');
      } else {
        const res = await api.post('/employees', payload);
        navigate(`/employees/${res.data.id}`);
      }
    } catch (err) {
      setServerError(extractError(err));
    }
  };

  if (loading) return <div className="empty">Loading…</div>;

  return (
    <>
      <div className="page-header">
        <h2>{isEdit ? 'Edit Employee' : 'Add Employee'}</h2>
        <button onClick={() => navigate('/employees')}>← Back to list</button>
      </div>

      {serverError && <div className="alert error">{serverError}</div>}
      {success && <div className="alert success">{success}</div>}

      <form className="card" onSubmit={handleSubmit(onSubmit)}>
        <div className="section-title">System Fields</div>
        <div className="form-grid">
          <div>
            <label>Name<span className="required-mark">*</span></label>
            <input {...register('name', { required: 'Name is required' })} />
            {errors.name && <div className="field-error">{errors.name.message}</div>}
          </div>
          <div>
            <label>Email<span className="required-mark">*</span></label>
            <input
              type="email"
              {...register('email', {
                required: 'Email is required',
                pattern: { value: /^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$/, message: 'Invalid email' },
              })}
            />
            {errors.email && <div className="field-error">{errors.email.message}</div>}
          </div>
          <div>
            <label>Phone</label>
            <input
              type="tel"
              {...register('phone', {
                validate: (v) => !v || /^\+?[0-9\s\-]{7,20}$/.test(v) || 'Invalid phone',
              })}
            />
            {errors.phone && <div className="field-error">{errors.phone.message}</div>}
          </div>
          <div>
            <label>Employee Code<span className="required-mark">*</span></label>
            <input {...register('employee_code', { required: 'Employee code is required' })} />
            {errors.employee_code && <div className="field-error">{errors.employee_code.message}</div>}
          </div>
          <div>
            <label>Department</label>
            <input {...register('department')} />
          </div>
          <div>
            <label>Designation</label>
            <input {...register('designation')} />
          </div>
          <div>
            <label>Date of Joining</label>
            <input type="date" {...register('date_of_joining')} />
          </div>
          <div>
            <label>Employment Type</label>
            <select {...register('employment_type')} defaultValue="">
              <option value="">— Select —</option>
              {EMPLOYMENT_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
            </select>
          </div>
          <div>
            <label>Status</label>
            <select {...register('status')} defaultValue="ACTIVE">
              {STATUSES.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
        </div>

        {fields.length > 0 && (
          <>
            <div className="section-title">Tenant Custom Fields</div>
            <div className="form-grid">
              {fields.map((f) => (
                <DynamicField key={f.id} field={f} register={register} errors={errors} />
              ))}
            </div>
          </>
        )}

        <div className="flex gap-8 mt-24">
          <button type="submit" className="primary" disabled={isSubmitting}>
            {isSubmitting ? 'Saving…' : (isEdit ? 'Save Changes' : 'Create Employee')}
          </button>
          <button type="button" onClick={() => navigate('/employees')}>Cancel</button>
        </div>
      </form>
    </>
  );
}
