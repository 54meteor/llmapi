import React, { useEffect, useState } from 'react';
import { Card, Grid, Header, Segment, Table, Progress, List } from 'semantic-ui-react';
import { API, showError, quota2string } from '../../helpers';

const Dashboard = () => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    let user = localStorage.getItem('user');
    if (user) {
      user = JSON.parse(user);
      setIsAdmin(user.role >= 10);
    }
    loadDashboard();
  }, []);

  const loadDashboard = async () => {
    setLoading(true);
    try {
      let res;
      if (isAdmin) {
        res = await API.get('/api/dashboard/admin');
      } else {
        res = await API.get('/api/user/usage');
      }
      const { success, message, data } = res.data;
      if (success) {
        setData(data);
      } else {
        showError(message);
      }
    } catch (e) {
      showError(e.message);
    }
    setLoading(false);
  };

  const renderAdminDashboard = () => {
    if (!data) return null;

    const { today, trend_7days, model_distribution, channel_health, top_users } = data;

    return (
      <>
        <Segment loading={loading}>
          <Header as='h3'>今日概况</Header>
          <Grid columns={4} stackable>
            <Grid.Column>
              <Card fluid color='blue'>
                <Card.Content>
                  <Card.Header>今日请求次数</Card.Header>
                  <Card.Description>{today?.request_count || 0}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='green'>
                <Card.Content>
                  <Card.Header>今日消耗配额</Card.Header>
                  <Card.Description>{quota2string(today?.quota_used || 0)}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='orange'>
                <Card.Content>
                  <Card.Header>活跃用户</Card.Header>
                  <Card.Description>{today?.active_users || 0}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='purple'>
                <Card.Content>
                  <Card.Header>活跃渠道</Card.Header>
                  <Card.Description>{today?.active_channels || 0}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid>
        </Segment>

        <Segment>
          <Header as='h3'>7天用量趋势</Header>
          {trend_7days && trend_7days.length > 0 ? (
            <Table basic='very'>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>日期</Table.HeaderCell>
                  <Table.HeaderCell>请求次数</Table.HeaderCell>
                  <Table.HeaderCell>消耗配额</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {trend_7days.map((item, idx) => (
                  <Table.Row key={idx}>
                    <Table.Cell>{item.day}</Table.Cell>
                    <Table.Cell>{item.request_count}</Table.Cell>
                    <Table.Cell>{quota2string(item.quota)}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          ) : (
            <p>暂无数据</p>
          )}
        </Segment>

        <Segment>
          <Header as='h3'>模型分布</Header>
          {model_distribution && model_distribution.length > 0 ? (
            <List>
              {model_distribution.map((item, idx) => (
                <List.Item key={idx}>
                  <List.Content>
                    <List.Header>{item.model}</List.Header>
                    <List.Description>
                      请求次数: {item.count} | 占比: {(item.ratio * 100).toFixed(1)}%
                    </List.Description>
                    <Progress percent={item.ratio * 100} indicating />
                  </List.Content>
                </List.Item>
              ))}
            </List>
          ) : (
            <p>暂无数据</p>
          )}
        </Segment>

        <Segment>
          <Header as='h3'>渠道健康状态</Header>
          {channel_health && channel_health.length > 0 ? (
            <Table basic='very'>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>渠道</Table.HeaderCell>
                  <Table.HeaderCell>成功率</Table.HeaderCell>
                  <Table.HeaderCell>平均响应时间</Table.HeaderCell>
                  <Table.HeaderCell>余额(USD)</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {channel_health.map((item, idx) => (
                  <Table.Row key={idx}>
                    <Table.Cell>{item.name}</Table.Cell>
                    <Table.Cell>{(item.success_rate * 100).toFixed(1)}%</Table.Cell>
                    <Table.Cell>{item.avg_response_time}ms</Table.Cell>
                    <Table.Cell>${item.balance.toFixed(2)}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          ) : (
            <p>暂无数据</p>
          )}
        </Segment>

        <Segment>
          <Header as='h3'>Top 10 用户用量</Header>
          {top_users && top_users.length > 0 ? (
            <Table basic='very'>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>用户名</Table.HeaderCell>
                  <Table.HeaderCell>消耗配额</Table.HeaderCell>
                  <Table.HeaderCell>请求次数</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {top_users.map((item, idx) => (
                  <Table.Row key={idx}>
                    <Table.Cell>{item.username}</Table.Cell>
                    <Table.Cell>{quota2string(item.quota_used)}</Table.Cell>
                    <Table.Cell>{item.request_count}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          ) : (
            <p>暂无数据</p>
          )}
        </Segment>
      </>
    );
  };

  const renderUserDashboard = () => {
    if (!data) return null;

    const { total_used, total_requests, quota_remaining, quota_percent, trend_7days } = data;

    return (
      <>
        <Segment loading={loading}>
          <Header as='h3'>我的用量</Header>
          <Grid columns={4} stackable>
            <Grid.Column>
              <Card fluid color='blue'>
                <Card.Content>
                  <Card.Header>累计消耗</Card.Header>
                  <Card.Description>{quota2string(total_used)}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='green'>
                <Card.Content>
                  <Card.Header>累计请求</Card.Header>
                  <Card.Description>{total_requests}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='orange'>
                <Card.Content>
                  <Card.Header>剩余配额</Card.Header>
                  <Card.Description>{quota2string(quota_remaining)}</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
            <Grid.Column>
              <Card fluid color='purple'>
                <Card.Content>
                  <Card.Header>使用百分比</Card.Header>
                  <Card.Description>{(quota_percent * 100).toFixed(1)}%</Card.Description>
                </Card.Content>
              </Card>
            </Grid.Column>
          </Grid>
          <Progress percent={quota_percent * 100} indicating style={{ marginTop: '1em' }} />
        </Segment>

        <Segment>
          <Header as='h3'>7天用量趋势</Header>
          {trend_7days && trend_7days.length > 0 ? (
            <Table basic='very'>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>日期</Table.HeaderCell>
                  <Table.HeaderCell>请求次数</Table.HeaderCell>
                  <Table.HeaderCell>消耗配额</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {trend_7days.map((item, idx) => (
                  <Table.Row key={idx}>
                    <Table.Cell>{item.day}</Table.Cell>
                    <Table.Cell>{item.request_count}</Table.Cell>
                    <Table.Cell>{quota2string(item.quota)}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          ) : (
            <p>暂无数据</p>
          )}
        </Segment>
      </>
    );
  };

  return (
    <>
      <Grid stackable>
        <Grid.Column width={16}>
          {isAdmin ? renderAdminDashboard() : renderUserDashboard()}
        </Grid.Column>
      </Grid>
    </>
  );
};

export default Dashboard;