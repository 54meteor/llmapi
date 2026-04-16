import PropTypes from 'prop-types';
import * as Yup from 'yup';
import { Formik } from 'formik';
import { useTheme } from '@mui/material/styles';
import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Divider,
  FormControl,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  FormHelperText
} from '@mui/material';

import { showSuccess, showError } from 'utils/common';
import { API } from 'utils/api';

const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  name: Yup.string().when('is_edit', {
    is: false,
    then: Yup.string().required('分组名称 不能为空'),
    otherwise: Yup.string()
  }),
  ratio: Yup.number().min(0, '费率倍率 不能小于 0')
});

const originInputs = {
  is_edit: false,
  name: '',
  ratio: 1
};

const EditModal = ({ open, groupId, onCancel, onOk }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState(originInputs);

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);

    let res;
    if (values.is_edit) {
      res = await API.put(`/api/group/${groupId}`, values);
    } else {
      res = await API.post(`/api/group/`, values);
    }
    const { success, message } = res.data;
    if (success) {
      if (values.is_edit) {
        showSuccess('分组更新成功！');
      } else {
        showSuccess('分组创建成功！');
      }
      setSubmitting(false);
      setStatus({ success: true });
      onOk(true);
    } else {
      showError(message);
      setErrors({ submit: message });
    }
  };

  const loadGroup = async () => {
    let res = await API.get(`/api/group/${groupId}`);
    const { success, message, data } = res.data;
    if (success) {
      data.is_edit = true;
      setInputs(data);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    if (groupId) {
      loadGroup().then();
    } else {
      setInputs(originInputs);
    }
  }, [groupId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={'sm'}>
      <DialogTitle sx={{ margin: '0px', fontWeight: 700, lineHeight: '1.55556', padding: '24px', fontSize: '1.125rem' }}>
        {groupId ? '编辑分组' : '新建分组'}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Formik initialValues={inputs} enableReinitialize validationSchema={validationSchema} onSubmit={submit}>
          {({ errors, handleBlur, handleChange, handleSubmit, touched, values, isSubmitting }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl fullWidth error={Boolean(touched.name && errors.name)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="group-name-label">分组名称</InputLabel>
                <OutlinedInput
                  id="group-name-label"
                  label="分组名称"
                  type="text"
                  value={values.name}
                  name="name"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: 'name' }}
                  aria-describedby="helper-text-group-name-label"
                  disabled={values.is_edit && values.name === 'default'}
                />
                {touched.name && errors.name && (
                  <FormHelperText error id="helper-text-group-name-label">
                    {errors.name}
                  </FormHelperText>
                )}
              </FormControl>

              <FormControl fullWidth error={Boolean(touched.ratio && errors.ratio)} sx={{ ...theme.typography.otherInput }}>
                <InputLabel htmlFor="group-ratio-label">费率倍率</InputLabel>
                <OutlinedInput
                  id="group-ratio-label"
                  label="费率倍率"
                  type="number"
                  value={values.ratio}
                  name="ratio"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ min: 0, step: 0.1 }}
                  startAdornment={<InputAdornment position="start">×</InputAdornment>}
                  aria-describedby="helper-text-group-ratio-label"
                />
                {touched.ratio && errors.ratio && (
                  <FormHelperText error id="helper-text-group-ratio-label">
                    {errors.ratio}
                  </FormHelperText>
                )}
              </FormControl>

              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button disableElevation disabled={isSubmitting} type="submit" variant="contained" color="primary">
                  提交
                </Button>
              </DialogActions>
            </form>
          )}
        </Formik>
      </DialogContent>
    </Dialog>
  );
};

export default EditModal;

EditModal.propTypes = {
  open: PropTypes.bool,
  groupId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func
};