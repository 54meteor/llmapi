import PropTypes from "prop-types";
import * as Yup from "yup";
import { Formik } from "formik";
import { useTheme } from "@mui/material/styles";
import { useState, useEffect } from "react";
import dayjs from "dayjs";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Divider,
  Alert,
  FormControl,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  Switch,
  FormHelperText,
  Select,
  MenuItem,
  Checkbox,
  ListItemText,
  List,
  ListItem,
  ListItemIcon,
  Box
} from "@mui/material";

import { AdapterDayjs } from "@mui/x-date-pickers/AdapterDayjs";
import { LocalizationProvider } from "@mui/x-date-pickers/LocalizationProvider";
import { DateTimePicker } from "@mui/x-date-pickers/DateTimePicker";
import { renderQuotaWithPrompt, showSuccess, showError, isAdmin } from "utils/common";
import { API } from "utils/api";
require("dayjs/locale/zh-cn");

const validationSchema = Yup.object().shape({
  is_edit: Yup.boolean(),
  name: Yup.string().required("名称 不能为空"),
  remain_quota: Yup.number().min(0, "必须大于等于0"),
  expired_time: Yup.number(),
  unlimited_quota: Yup.boolean(),
});

const originInputs = {
  is_edit: false,
  name: "",
  remain_quota: 0,
  expired_time: -1,
  unlimited_quota: false,
  user_id: 0,
  channel_ids: []
};

const EditModal = ({ open, tokenId, onCancel, onOk }) => {
  const theme = useTheme();
  const [inputs, setInputs] = useState(originInputs);
  const [users, setUsers] = useState([]);
  const [channels, setChannels] = useState([]);
  const [selectedChannels, setSelectedChannels] = useState([]);
  const isAdminUser = isAdmin();

  const loadUsers = async () => {
    if (!isAdminUser) return;
    const res = await API.get('/api/user/');
    const { success, data } = res.data;
    if (success) {
      setUsers(data);
    }
  };

  const loadChannels = async () => {
    const res = await API.get('/api/channel/');
    const { success, data } = res.data;
    if (success) {
      setChannels(data);
    }
  };

  const loadTokenChannels = async () => {
    if (!tokenId) return;
    const res = await API.get(`/api/token-channel/${tokenId}`);
    const { success, data } = res.data;
    if (success) {
      const boundChannelIds = data.map(tc => tc.channel_id);
      setSelectedChannels(boundChannelIds);
    }
  };

  useEffect(() => {
    if (open) {
      loadUsers();
      loadChannels();
      if (tokenId) {
        loadTokenChannels();
      }
    }
  }, [open, tokenId]);

  const submit = async (values, { setErrors, setStatus, setSubmitting }) => {
    setSubmitting(true);

    values.remain_quota = parseInt(values.remain_quota);
    let res;
    if (values.is_edit) {
      res = await API.put(`/api/token/`, { ...values, id: parseInt(tokenId) });
    } else {
      const queryParams = values.user_id ? `?user_id=${values.user_id}` : '';
      res = await API.post(`/api/token/${queryParams}`, values);
    }
    const { success, message } = res.data;
    if (success) {
      if (values.is_edit) {
        showSuccess("令牌更新成功！");
      } else {
        showSuccess("令牌创建成功，请在列表页面点击复制获取令牌！");
      }
      if (!values.is_edit && selectedChannels.length > 0) {
        const newTokenId = tokenId || res.data.data?.id;
        if (newTokenId) {
          for (const channelId of selectedChannels) {
            await API.post('/api/token-channel/', { token_id: newTokenId, channel_id: channelId });
          }
        }
      }
      setSubmitting(false);
      setStatus({ success: true });
      onOk(true);
    } else {
      showError(message);
      setErrors({ submit: message });
    }
  };

  const loadToken = async () => {
    let res = await API.get(`/api/token/${tokenId}`);
    const { success, message, data } = res.data;
    if (success) {
      data.is_edit = true;
      data.user_id = data.user_id || 0;
      setInputs(data);
      const tokenChannelsRes = await API.get(`/api/token-channel/${tokenId}`);
      const { success: tcSuccess, data: tcData } = tokenChannelsRes.data;
      if (tcSuccess) {
        setSelectedChannels(tcData.map(tc => tc.channel_id));
      }
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    if (tokenId) {
      loadToken().then();
    }
  }, [tokenId]);

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth={"md"}>
      <DialogTitle
        sx={{
          margin: "0px",
          fontWeight: 700,
          lineHeight: "1.55556",
          padding: "24px",
          fontSize: "1.125rem",
        }}
      >
        {tokenId ? "编辑Token" : "新建Token"}
      </DialogTitle>
      <Divider />
      <DialogContent>
        <Alert severity="info">
          注意，令牌的额度仅用于限制令牌本身的最大额度使用量，实际的使用受到账户的剩余额度限制。
        </Alert>
        <Formik
          initialValues={inputs}
          enableReinitialize
          validationSchema={validationSchema}
          onSubmit={submit}
        >
          {({
            errors,
            handleBlur,
            handleChange,
            handleSubmit,
            touched,
            values,
            setFieldError,
            setFieldValue,
            isSubmitting,
          }) => (
            <form noValidate onSubmit={handleSubmit}>
              <FormControl
                fullWidth
                error={Boolean(touched.name && errors.name)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-name-label">名称</InputLabel>
                <OutlinedInput
                  id="channel-name-label"
                  label="名称"
                  type="text"
                  value={values.name}
                  name="name"
                  onBlur={handleBlur}
                  onChange={handleChange}
                  inputProps={{ autoComplete: "name" }}
                  aria-describedby="helper-text-channel-name-label"
                />
                {touched.name && errors.name && (
                  <FormHelperText error id="helper-tex-channel-name-label">
                    {errors.name}
                  </FormHelperText>
                )}
              </FormControl>
              {isAdminUser && (
                <FormControl fullWidth sx={{ ...theme.typography.otherInput }}>
                  <InputLabel>归属用户</InputLabel>
                  <Select
                    label="归属用户"
                    name="user_id"
                    value={values.user_id || 0}
                    onChange={(e) => {
                      setFieldValue('user_id', e.target.value);
                    }}
                  >
                    <MenuItem value={0}>请选择用户</MenuItem>
                    {users.map((user) => (
                      <MenuItem key={user.id} value={user.id}>{user.username}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              )}
              {values.expired_time !== -1 && (
                <FormControl
                  fullWidth
                  error={Boolean(touched.expired_time && errors.expired_time)}
                  sx={{ ...theme.typography.otherInput }}
                >
                  <LocalizationProvider
                    dateAdapter={AdapterDayjs}
                    adapterLocale={"zh-cn"}
                  >
                    <DateTimePicker
                      label="过期时间"
                      ampm={false}
                      value={dayjs.unix(values.expired_time)}
                      onError={(newError) => {
                        if (newError === null) {
                          setFieldError("expired_time", null);
                        } else {
                          setFieldError("expired_time", "无效的日期");
                        }
                      }}
                      onChange={(newValue) => {
                        setFieldValue("expired_time", newValue.unix());
                      }}
                      slotProps={{
                        actionBar: {
                          actions: ["today", "accept"],
                        },
                      }}
                    />
                  </LocalizationProvider>
                  {errors.expired_time && (
                    <FormHelperText
                      error
                      id="helper-tex-channel-expired_time-label"
                    >
                      {errors.expired_time}
                    </FormHelperText>
                  )}
                </FormControl>
              )}
              <Switch
                checked={values.expired_time === -1}
                onClick={() => {
                  if (values.expired_time === -1) {
                    setFieldValue(
                      "expired_time",
                      Math.floor(Date.now() / 1000)
                    );
                  } else {
                    setFieldValue("expired_time", -1);
                  }
                }}
              />{" "}
              永不过期
              <FormControl
                fullWidth
                error={Boolean(touched.remain_quota && errors.remain_quota)}
                sx={{ ...theme.typography.otherInput }}
              >
                <InputLabel htmlFor="channel-remain_quota-label">
                  额度
                </InputLabel>
                <OutlinedInput
                  id="channel-remain_quota-label"
                  label="额度"
                  type="number"
                  value={values.remain_quota}
                  name="remain_quota"
                  endAdornment={
                    <InputAdornment position="end">
                      {renderQuotaWithPrompt(values.remain_quota)}
                    </InputAdornment>
                  }
                  onBlur={handleBlur}
                  onChange={handleChange}
                  aria-describedby="helper-text-channel-remain_quota-label"
                  disabled={values.unlimited_quota}
                />

                {touched.remain_quota && errors.remain_quota && (
                  <FormHelperText
                    error
                    id="helper-tex-channel-remain_quota-label"
                  >
                    {errors.remain_quota}
                  </FormHelperText>
                )}
              </FormControl>
              <Switch
                checked={values.unlimited_quota === true}
                onClick={() => {
                  setFieldValue("unlimited_quota", !values.unlimited_quota);
                }}
              />{" "}
              无限额度
              <FormControl fullWidth sx={{ ...theme.typography.otherInput, mt: 2 }}>
                <InputLabel>绑定渠道</InputLabel>
                <Select
                  multiple
                  label="绑定渠道"
                  value={selectedChannels}
                  onChange={(e) => {
                    setSelectedChannels(e.target.value);
                  }}
                  renderValue={(selected) => {
                    const selectedNames = channels
                      .filter(ch => selected.includes(ch.id))
                      .map(ch => ch.name)
                      .join(', ');
                    return selectedNames || '请选择渠道';
                  }}
                >
                  {channels.map((channel) => (
                    <MenuItem key={channel.id} value={channel.id}>
                      <Checkbox checked={selectedChannels.indexOf(channel.id) !== -1} />
                      <ListItemText primary={channel.name} />
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <DialogActions>
                <Button onClick={onCancel}>取消</Button>
                <Button
                  disableElevation
                  disabled={isSubmitting}
                  type="submit"
                  variant="contained"
                  color="primary"
                >
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
  tokenId: PropTypes.number,
  onCancel: PropTypes.func,
  onOk: PropTypes.func,
};
