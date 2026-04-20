import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Message, Segment, Table, Modal, Input, Dropdown } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess, timestamp2string, isAdmin } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';
import ChannelSelector from '../../components/ChannelSelector';

const EditToken = () => {
  const params = useParams();
  const tokenId = params.id;
  const isEdit = tokenId !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const originInputs = {
    name: '',
    remain_quota: isEdit ? 0 : 500000,
    expired_time: -1,
    unlimited_quota: false,
    switch_threshold: 10,
    switch_threshold_type: 'percent',
    alert_threshold: 5,
    alert_threshold_type: 'percent',
    smart_channel_enabled: true,
    user_id: isEdit ? 0 : ''
  };
  const [inputs, setInputs] = useState(originInputs);
  const [tokenChannels, setTokenChannels] = useState([]);
  const [editingIds, setEditingIds] = useState(new Set());
  const [editingValues, setEditingValues] = useState({});
  const [showChannelSelector, setShowChannelSelector] = useState(false);
  const [users, setUsers] = useState([]);
  const [userOptions, setUserOptions] = useState([]);
  const { name, remain_quota, expired_time, unlimited_quota } = inputs;
  const navigate = useNavigate();
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
const handleCancel = () => {
    navigate("/token");
  };

  const handleDeleteChannel = async (id) => {
    if (!window.confirm('确定要解绑该渠道吗？')) return;
    let res = await API.delete(`/api/token-channel/${id}`);
    const { success, message } = res.data;
    if (success) {
      showSuccess('渠道解绑成功！');
      loadTokenChannels();
    } else {
      showError(message);
    }
  };

  const startEdit = (tc) => {
    setEditingIds(new Set([...editingIds, tc.id]));
    setEditingValues({
      ...editingValues,
      [tc.id]: {
        priority: tc.priority,
        quota_limit: tc.quota_limit
      }
    });
  };

  const cancelEdit = (id) => {
    const newIds = new Set(editingIds);
    newIds.delete(id);
    setEditingIds(newIds);
    const newValues = { ...editingValues };
    delete newValues[id];
    setEditingValues(newValues);
  };

  const saveEdit = async (tc) => {
    const values = editingValues[tc.id];
    if (!values) return;
    let res = await API.put(`/api/token-channel/${tc.id}`, {
      priority: parseInt(values.priority) || 1,
      quota_limit: parseInt(values.quota_limit) || 0
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess('渠道配置更新成功！');
      cancelEdit(tc.id);
      loadTokenChannels();
    } else {
      showError(message);
    }
  };

  const setExpiredTime = (month, day, hour, minute) => {
    let now = new Date();
    let timestamp = now.getTime() / 1000;
    let seconds = month * 30 * 24 * 60 * 60;
    seconds += day * 24 * 60 * 60;
    seconds += hour * 60 * 60;
    seconds += minute * 60;
    if (seconds !== 0) {
      timestamp += seconds;
      setInputs({ ...inputs, expired_time: timestamp2string(timestamp) });
    } else {
      setInputs({ ...inputs, expired_time: -1 });
    }
  };

  const setUnlimitedQuota = () => {
    setInputs({ ...inputs, unlimited_quota: !unlimited_quota });
  };

  const loadToken = async () => {
    let res = await API.get(`/api/token/${tokenId}`);
    const { success, message, data } = res.data;
    if (success) {
      if (data.expired_time !== -1) {
        data.expired_time = timestamp2string(data.expired_time);
      }
      if (data.switch_threshold === undefined) {
        data.switch_threshold = 10;
        data.switch_threshold_type = 'percent';
        data.alert_threshold = 5;
        data.alert_threshold_type = 'percent';
        data.smart_channel_enabled = true;
      }
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const loadTokenChannels = async () => {
    let res = await API.get(`/api/token-channel/${tokenId}`);
    const { success, message, data } = res.data;
    if (success) {
      setTokenChannels(data || []);
    } else {
      showError(message);
    }
  };

  const loadUsers = async () => {
    if (!isAdmin()) return;
    try {
      let res = await API.get(`/api/user/`);
      if (res.data.success) {
        setUsers(res.data.data || []);
        setUserOptions(
          (res.data.data || []).map((user) => ({
            key: user.id,
            text: user.username,
            value: user.id,
          }))
        );
      }
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    if (isEdit) {
      loadToken().then();
      loadTokenChannels().then();
    }
    loadUsers().then();
  }, []);

  const submit = async () => {
    if (!isEdit && inputs.name === '') return;
    let localInputs = inputs;
    localInputs.remain_quota = parseInt(localInputs.remain_quota);
    if (localInputs.expired_time !== -1) {
      let time = Date.parse(localInputs.expired_time);
      if (isNaN(time)) {
        showError('过期时间格式错误！');
        return;
      }
      localInputs.expired_time = Math.ceil(time / 1000);
    }
    let res;
    if (isEdit) {
      res = await API.put(`/api/token/`, { ...localInputs, id: parseInt(tokenId) });
    } else {
      // Pass user_id as query parameter for admin
      if (inputs.user_id && isAdmin()) {
        res = await API.post(`/api/token/?user_id=${inputs.user_id}`, localInputs);
      } else {
        res = await API.post(`/api/token/`, localInputs);
      }
    }
    const { success, message } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess('令牌更新成功！');
      } else {
        showSuccess('令牌创建成功，请在列表页面点击复制获取令牌！');
        setInputs(originInputs);
      }
    } else {
      showError(message);
    }
  };

  return (
    <>
      <Segment loading={loading}>
        <Header as='h3'>{isEdit ? '更新令牌信息' : '创建新的令牌'}</Header>
        <Form autoComplete='new-password'>
          <Form.Field>
            <Form.Input
              label='名称'
              name='name'
              placeholder={'请输入名称'}
              onChange={handleInputChange}
              value={name}
              autoComplete='new-password'
              required={!isEdit}
            />
          </Form.Field>
          {isAdmin() && (
            <Form.Field>
              <label>所属用户</label>
              <Dropdown
                placeholder='选择用户'
                fluid
                selection
                clearable
                search
                options={userOptions}
                value={inputs.user_id || ''}
                onChange={(_, { value }) => {
                  setInputs({ ...inputs, user_id: value || '' });
                }}
              />
            </Form.Field>
          )}
          <Form.Field>
            <Form.Input
              label='过期时间'
              name='expired_time'
              placeholder={'请输入过期时间，格式为 yyyy-MM-dd HH:mm:ss，-1 表示无限制'}
              onChange={handleInputChange}
              value={expired_time}
              autoComplete='new-password'
              type='datetime-local'
            />
          </Form.Field>
          <div style={{ lineHeight: '40px' }}>
            <Button type={'button'} onClick={() => {
              setExpiredTime(0, 0, 0, 0);
            }}>永不过期</Button>
            <Button type={'button'} onClick={() => {
              setExpiredTime(1, 0, 0, 0);
            }}>一个月后过期</Button>
            <Button type={'button'} onClick={() => {
              setExpiredTime(0, 1, 0, 0);
            }}>一天后过期</Button>
            <Button type={'button'} onClick={() => {
              setExpiredTime(0, 0, 1, 0);
            }}>一小时后过期</Button>
            <Button type={'button'} onClick={() => {
              setExpiredTime(0, 0, 0, 1);
            }}>一分钟后过期</Button>
          </div>
          <Message>注意，令牌的额度仅用于限制令牌本身的最大额度使用量，实际的使用受到账户的剩余额度限制。</Message>
          <Form.Field>
            <Form.Input
              label={`额度${renderQuotaWithPrompt(remain_quota)}`}
              name='remain_quota'
              placeholder={'请输入额度'}
              onChange={handleInputChange}
              value={remain_quota}
              autoComplete='new-password'
              type='number'
              disabled={unlimited_quota}
            />
          </Form.Field>
          <Button type={'button'} onClick={() => {
            setUnlimitedQuota();
          }}>{unlimited_quota ? '取消无限额度' : '设为无限额度'}</Button>
          <Segment>
            <Header as='h4'>渠道路由配置</Header>
            <Form.Group widths='equal'>
              <Form.Input
                label='切换阈值'
                placeholder='10'
                name='switch_threshold'
                type='number'
                value={inputs.switch_threshold || 10}
                onChange={handleInputChange}
              />
              <Form.Select
                label='类型'
                name='switch_threshold_type'
                options={[{key:'percent',text:'%',value:'percent'},{key:'absolute',text:'绝对值',value:'absolute'}]}
                onChange={(e,{value}) => setInputs({...inputs, switch_threshold_type: value})}
                value={inputs.switch_threshold_type || 'percent'}
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.Input
                label='报警阈值'
                placeholder='5'
                name='alert_threshold'
                type='number'
                value={inputs.alert_threshold || 5}
                onChange={handleInputChange}
              />
              <Form.Select
                label='类型'
                name='alert_threshold_type'
                options={[{key:'percent',text:'%',value:'percent'},{key:'absolute',text:'绝对值',value:'absolute'}]}
                onChange={(e,{value}) => setInputs({...inputs, alert_threshold_type: value})}
                value={inputs.alert_threshold_type || 'percent'}
              />
            </Form.Group>
            <Form.Checkbox
              label='启用智能渠道切换'
              checked={inputs.smart_channel_enabled !== false}
              onChange={(e, { checked }) => setInputs({...inputs, smart_channel_enabled: checked})}
            />
          </Segment>
          {isEdit && (
            <Segment>
              <Header as='h4'>已绑定渠道</Header>
              <Button color='green' onClick={() => setShowChannelSelector(true)}>
                绑定新渠道
              </Button>
              {tokenChannels.length > 0 ? (
                <Table>
                  <Table.Header>
                    <Table.Row>
                      <Table.HeaderCell>优先级</Table.HeaderCell>
                      <Table.HeaderCell>渠道名称</Table.HeaderCell>
                      <Table.HeaderCell>渠道类型</Table.HeaderCell>
                      <Table.HeaderCell>额度上限</Table.HeaderCell>
                      <Table.HeaderCell>已用</Table.HeaderCell>
                      <Table.HeaderCell>剩余</Table.HeaderCell>
                      <Table.HeaderCell>操作</Table.HeaderCell>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {tokenChannels.map(tc => (
                      <Table.Row key={tc.id}>
                        <Table.Cell>
                          {editingIds.has(tc.id) ? (
                            <Form.Input
                              size='mini'
                              type='number'
                              value={editingValues[tc.id]?.priority ?? tc.priority}
                              onChange={(e, { value }) => setEditingValues({
                                ...editingValues,
                                [tc.id]: { ...editingValues[tc.id], priority: value }
                              })}
                            />
                          ) : tc.priority}
                        </Table.Cell>
                        <Table.Cell>{tc.channel_name}</Table.Cell>
                        <Table.Cell>{tc.channel_type}</Table.Cell>
                        <Table.Cell>
                          {editingIds.has(tc.id) ? (
                            <Form.Input
                              size='mini'
                              type='number'
                              placeholder='0 表示不限'
                              value={editingValues[tc.id]?.quota_limit ?? tc.quota_limit}
                              onChange={(e, { value }) => setEditingValues({
                                ...editingValues,
                                [tc.id]: { ...editingValues[tc.id], quota_limit: value }
                              })}
                            />
                          ) : (tc.quota_limit || '不限')}
                        </Table.Cell>
                        <Table.Cell>{tc.used_quota}</Table.Cell>
                        <Table.Cell>{tc.remain_quota} ({tc.remain_percent}%)</Table.Cell>
                        <Table.Cell>
                          {editingIds.has(tc.id) ? (
                            <>
                              <Button size='small' primary onClick={() => saveEdit(tc)}>保存</Button>
                              <Button size='small' onClick={() => cancelEdit(tc.id)}>取消</Button>
                            </>
                          ) : (
                            <>
                              <Button size='small' onClick={() => startEdit(tc)}>编辑</Button>
                              <Button size='small' onClick={() => handleDeleteChannel(tc.id)}>解绑</Button>
                            </>
                          )}
                        </Table.Cell>
                      </Table.Row>
                    ))}
                  </Table.Body>
                </Table>
              ) : <Message>暂未绑定渠道</Message>}
            </Segment>
          )}
          <ChannelSelector
            open={showChannelSelector}
            onClose={() => setShowChannelSelector(false)}
            onSuccess={() => {
              loadTokenChannels();
              setShowChannelSelector(false);
            }}
            tokenId={parseInt(tokenId)}
          />
          <Button floated='right' positive onClick={submit}>提交</Button>
          <Button floated='right' onClick={handleCancel}>取消</Button>
        </Form>
      </Segment>
    </>
  );
};

export default EditToken;
